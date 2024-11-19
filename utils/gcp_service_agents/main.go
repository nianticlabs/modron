package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

const (
	serviceAgentsUrl           = "https://cloud.google.com/iam/docs/service-agents"
	firebaseServiceAccountsUrl = "https://firebase.google.com/support/guides/service-accounts"
)

var (
	log                  = logrus.StandardLogger()
	undocumentedProjects = map[string]struct{}{
		"appsheet-prod-service-accounts": {},
		"cloud-ml.google.com":            {},
		"cloud-cdn-fill":                 {},
		"gae-api-prod.google.com":        {},
	}
)

var args struct {
	OutputFile string `arg:"-o,required" help:"Output file to write the JSON to"`
}

func main() {
	arg.MustParse(&args)

	f, err := os.Create(args.OutputFile)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer f.Close()

	if err := generateServiceAgents(f); err != nil {
		log.Fatalf("failed to generate service agents: %v", err)
	}
}

type ServiceAgentProjects struct {
	Projects []string `json:"projects"`
}

func generateServiceAgents(f io.Writer) error {
	projects := make(map[string]struct{})
	if err := scrapeServiceAgentsPage(projects); err != nil {
		return err
	}

	if err := scrapeFirebaseServiceAccountsPage(projects); err != nil {
		return err
	}

	for _, undocumentedProjects := range maps.Keys(undocumentedProjects) {
		projects[undocumentedProjects] = struct{}{}
	}

	projectIDs := maps.Keys(projects)
	sort.Strings(projectIDs)
	toWrite := ServiceAgentProjects{Projects: projectIDs}
	if err := json.NewEncoder(f).Encode(toWrite); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}
	return nil
}

func scrapeFirebaseServiceAccountsPage(projects map[string]struct{}) error {
	doc, err := getGoQueryDocument(firebaseServiceAccountsUrl)
	if err != nil {
		return err
	}
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		svcAccount := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		svcAccountProject := getServiceAccountProject(svcAccount)
		if len(svcAccountProject) == 0 {
			log.Warnf("Service account project not found for %s", strings.ReplaceAll(svcAccount, "\n", " "))
			return
		}
		for _, saProjectID := range svcAccountProject {
			projects[saProjectID] = struct{}{}
		}
	})
	return nil
}

func getGoQueryDocument(pageURL string) (*goquery.Document, error) {
	page, err := scrapePage(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape page: %v", err)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(page))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}
	return doc, nil
}

func scrapeServiceAgentsPage(projects map[string]struct{}) error {
	doc, err := getGoQueryDocument(serviceAgentsUrl)
	if err != nil {
		return err
	}
	doc.Find("#service-agents tbody tr").Each(func(i int, s *goquery.Selection) {
		serviceAgent := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		svcAccountProject := getServiceAccountProject(serviceAgent)
		if len(svcAccountProject) == 0 {
			log.Warnf("Service agent project not found for %s", strings.ReplaceAll(serviceAgent, "\n", " "))
			return
		}
		log.Infof("Service agent: %s", svcAccountProject)
		for _, saProjectID := range svcAccountProject {
			projects[saProjectID] = struct{}{}
		}
	})
	return nil
}

var saAccountProjectRegex = regexp.MustCompile("\\S+@([A-z0-9-]+)\\.iam\\.gserviceaccount\\.com")

func getServiceAccountProject(agent string) []string {
	projectsMap := make(map[string]struct{})
	matches := saAccountProjectRegex.FindAllStringSubmatch(agent, -1)
	for _, match := range matches {
		if len(match) != 2 {
			log.Warnf("failed to extract project from service agent: %s", agent)
		} else {
			if match[1] == "project-id" || match[1] == "project-name" {
				continue
			}
			projectsMap[match[1]] = struct{}{}
		}
	}
	return maps.Keys(projectsMap)
}

func scrapePage(pageUrl string) (string, error) {
	req, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got status code %d", res.StatusCode)
	}
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, res.Body); err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	return buffer.String(), nil
}
