package springio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/eslam-allam/spring-initializer-go/models/metadata"
)


type metaFieldType string

var client http.Client = http.Client{
    Timeout: constants.DownloadTimeoutSeconds * time.Second,
}

const (
	TEXT          metaFieldType = "text"
	SINGLE_SELECT metaFieldType = "single-select"
	MULTI_SELECT  metaFieldType = "heirarchal-multi-select"
	ACTION        metaFieldType = "action"
)

type metaField struct {
	Id          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        metaFieldType `json:"type"`
	Default     string        `json:"default"`
	Action      string        `json:"action"`
	Values      []metaField   `json:"values"`
}

type SpringInitMeta struct {
	ArtifactId   metaField `json:"artifactId"`
	BootVersion  metaField `json:"bootVersion"`
	Dependencies metaField `json:"dependencies"`
	Description  metaField `json:"description"`
	GroupId      metaField `json:"groupId"`
	JavaVersion  metaField `json:"javaVersion"`
	Language     metaField `json:"language"`
	Name         metaField `json:"name"`
	PackageName  metaField `json:"packageName"`
	Packaging    metaField `json:"packaging"`
	Type         metaField `json:"type"`
	Version      metaField `json:"version"`
}

func GetMeta() (SpringInitMeta, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", constants.SpringUrl, nil)
	req.Header.Set("Accept", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return SpringInitMeta{}, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return SpringInitMeta{}, err
	}

	var responseObject SpringInitMeta
	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		return SpringInitMeta{}, err
	}
	return responseObject, nil
}

func DownloadGeneratedZip(url string, filepath string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error downloading file: %s, %s", resp.Status, body)
	}
	defer resp.Body.Close()
	// Create the output file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	// Copy the response body to the output file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func GenerateDownloadRequest(action, project, language, springBootVersion,
	packaging, javaVersion string, dependencies []string, metadata []metadata.FieldValue,
) (*url.URL, error) {
	form := url.Values{}

	for _, m := range metadata {
		form.Add(m.Id, m.Value)
	}

	form.Add("type", project)
	form.Add("language", language)
	form.Add("bootVersion", springBootVersion)
	form.Add("packaging", packaging)
	form.Add("javaVersion", javaVersion)

	for _, d := range dependencies {
		form.Add("dependencies", d)
	}

	url, error := url.Parse(fmt.Sprintf("%s?%s", constants.SpringUrl, form.Encode()))

	if error != nil {
		return url, error
	}

	url = url.JoinPath(action)

	return url, nil
}
