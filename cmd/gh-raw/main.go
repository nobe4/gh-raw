package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func startHttpServer(wg *sync.WaitGroup) (*http.Server, int) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	srv := &http.Server{
		Addr: ":0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request received %#v", r)
			formatRequest(*r)
		}),
		// XXX: this changes nothing
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	}

	go func() {
		defer wg.Done()

		log.Printf("listening on :%d\n", port)
		// XXX: certbot won't generate a cert for localhost, self-signed cert
		// also doesn't work.
		// if err := srv.ServeTLS(listener, "/tmp/cert/localhost.crt", "/tmp/cert/localhost.key"); err != http.ErrServerClosed {
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	return srv, port
}

const curlTemplate = `
curl -X {{ .Method }} \
  {{- /* {{ if .TLS }}https://{{ else }}http://{{ end }}{{ .Host }}{{ .URL.Scheme }} \ */}}
  https://api.github.com{{ .URL }} \
  {{ range $key, $values := .Header -}}
    {{- if eq $key "Authorization" -}}
      -H "{{ $key }}: token $GITHUB_TOKEN" \
    {{- else -}}
       {{- range $values -}}
         -H "{{ $key }}: {{ . }}" \
       {{- end -}}
    {{ end }}
  {{ end }}
`

func formatRequest(r http.Request) {
	t, err := template.New("test").Parse(curlTemplate)
	if err != nil {
		panic(err)
	}

	err = t.Execute(os.Stdout, r)
	if err != nil {
		panic(err)
	}
}

func main() {
	log.Printf("main: starting HTTP server")

	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv, port := startHttpServer(httpServerExitDone)

	log.Printf("main: serving until canceled or interrupted")
	log.Printf("args: %v", os.Args)

	go func() {
		defer func() { done <- syscall.SIGTERM }()

		// XXX: binary built with https://github.com/nobe4/cli/pull/1
		cmd := exec.Command("/tmp/gh")

		// XXX: This fails because gh won't use a non-https host.
		// cc https://github.com/cli/cli/issues/8640
		cmd.Env = []string{
			fmt.Sprintf("GH_ENTERPRISE_TOKEN=%s", os.Getenv("GH_TOKEN")),
			// "GH_DEBUG=api",
			fmt.Sprintf("GH_HOST=github.localhost:%d", port),
		}
		cmd.Args = append(cmd.Args, os.Args[1:]...)

		// log.Printf("cmd: %#v", cmd)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
	}()

	signal := <-done
	log.Printf("main: got signal: %v", signal)

	log.Printf("main: stopping HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Printf("main: done. exiting")
}
