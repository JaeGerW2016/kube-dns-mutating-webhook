package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"k8s.io/klog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)


func main() {
	var parameters WhSvrParameters
	//get cli parameters
	flag.IntVar(&parameters.port, "port", 443, "webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.StringVar(&parameters.dnsCfgFile, "kubednsCfgFile", "/etc/webhook/config/kubednsCfgFile.yaml", "File containing the mutation configuration.")
	flag.Parse()

	dnsConfig, err := loadConfig(parameters.dnsCfgFile)
	if err != nil {
		klog.Errorf("Failed to load configuration: %v", err)
	}

	pair, err := tls.LoadX509KeyPair(parameters.certFile, parameters.keyFile)
	if err != nil {
		klog.Errorf("Failed to loadx509keypair: %v", err)
	}

	whsvr := &WebhookServer{
		poddnsConfig: dnsConfig,
		server: &http.Server{
			Addr:      fmt.Sprintf(":%v", parameters.port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsvr.serve)
	whsvr.server.Handler = mux
	klog.Infof("Starting webhook server ...")

	go func() {
		if err := whsvr.server.ListenAndServeTLS("", ""); err != nil {
			klog.Errorf("Failed to listen and server webhook server:%v", err)
		}
	}()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	klog.Infof("Got OS Shutdown signal,shutting down webhook server gracefully ...")
	_ = whsvr.server.Shutdown(context.Background())
}
