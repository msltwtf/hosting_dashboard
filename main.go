package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*

var content embed.FS

type ServerData struct {
	Name      string
	PublicNet hcloud.ServerPublicNet
	Labels    map[string]string
}

type NamespaceData struct {
	Namespace string
	Servers   []*hcloud.Server
}

type toolConfig struct {
	Namespaces map[string]string
}

var config toolConfig
var configfile string

func (c *toolConfig) Load(filename string) error {
	if strings.HasSuffix(filename, ".json") {
		data, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err.Error())
		}
		return json.Unmarshal(data, &config)
	} else if strings.HasSuffix(filename, ".yml") || strings.HasSuffix(filename, ".yaml") {
		data, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err.Error())
		}
		return yaml.Unmarshal(data, &config)
	} else {
		return errors.New("invalid file format! please use json or yaml")
	}
}

func getFromAll(w http.ResponseWriter, r *http.Request) {
	var AllData []NamespaceData
	for namespace, token := range config.Namespaces {
		client := hcloud.NewClient(hcloud.WithToken(token))
		servers, err := client.Server.All(context.Background())
		if err != nil {
			log.Fatal(err.Error())
		}
		data := NamespaceData{
			Namespace: namespace,
			Servers:   servers,
		}
		AllData = append(AllData, data)
	}

	t, err := template.ParseFS(content, "templates/servers.gohtml")
	if err != nil {
		log.Fatal(err.Error())
	}
	d := struct {
		AllData []NamespaceData
	}{
		AllData: AllData,
	}
	err = t.Execute(w, d)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	flag.StringVar(&configfile, "c", "config.yml", "Path to config file")
	flag.Parse()
	log.Println("Starting Hosting Dashboard")
	log.Println("Loading Config")
	if err := config.Load(configfile); err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Loaded config")

	http.HandleFunc("/", getFromAll)
	http.ListenAndServe(":8080", nil)
}
