package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"text/template"
	"time"
)

func main() {
	backend := flag.String("backend", "http://endpoint:8080/api.json", "check endpoint")
	templatePath := flag.String("template", "", "config template")
	outputPath := flag.String("output", "", "config output path")
	flag.Parse()

	// コマンド
	cmd := flag.Args()
	if len(cmd) > 0 && cmd[0] == "--" {
		cmd = cmd[1:]
	}
	if len(cmd) < 1 {
		log.Fatal("no commands specified")
	}
	log.Printf("cmd: %+v\n", cmd)

	// 初期実行
	apiResp, err := callApi(*backend)
	if err != nil {
		log.Fatal(err)
	}
	if err := writeConfig(*outputPath, *apiResp, *templatePath); err != nil {
		log.Fatal(err)
	}

	// ここから、起動させたり終了したりのコード
	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}

	childProcess := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	childProcess.Stderr = os.Stderr
	childProcess.Stdout = os.Stdout

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := childProcess.Start()
		if err != nil {
			log.Printf("%s\n", err)
			cancel()
		}
		log.Printf("Waiting for command to finish...")
		err = childProcess.Wait()
		log.Printf("Command finished with error: %v", err)
		cancel()
	}()

	// 起動待ち...
	// TODO タイムアウトはいる?
	for {
		if childProcess.Process != nil {
			log.Printf("pid: %d\n", childProcess.Process.Pid)
			break
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Printf("done.\n")
				return
			case <-time.After(1 * time.Second):
				// ポーリングする
				newApiResp, err := callApi(*backend)
				if err != nil {
					log.Fatal(err)
				}
				if reflect.DeepEqual(*apiResp, *newApiResp) {
					// do nothing
				} else {
					log.Printf("update servers!\n")
					apiResp = newApiResp
					if err := writeConfig(*outputPath, *apiResp, *templatePath); err != nil {
						log.Fatal(err)
					}
					sendSIGHUP(*childProcess)
				}
			}
		}
	}()

	s := <-sigChan
	switch s {
	case syscall.SIGTERM:
		log.Println("!!! SIGTERM detected !!!!")
		cancel()
		wg.Wait()
		log.Println("successful termination")
	default:
		log.Println("unexpected signal")
	}
}

func writeConfig(outputPath string, apiResp ApiResp, templatePath string) error {
	f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	t := template.Must(template.ParseFiles(templatePath))
	if err := t.Execute(f, apiResp); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func sendSIGHUP(childProcess exec.Cmd) error {
	log.Printf("kill: %d\n", childProcess.Process.Pid)

	// send SIGHUP
	err := childProcess.Process.Signal(syscall.SIGHUP)
	if err != nil {
		return err
	}
	return nil
}
