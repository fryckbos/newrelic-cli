package install

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"gopkg.in/yaml.v2"
)

type recipeFile struct {
	Description    string                 `yaml:"description"`
	InputVars      []variableConfig       `yaml:"inputVars"`
	Install        map[string]interface{} `yaml:"install"`
	InstallTargets []recipeInstallTarget  `yaml:"installTargets"`
	Keywords       []string               `yaml:"keywords"`
	LogMatch       []logMatch             `yaml:"logMatch"`
	Name           string                 `yaml:"name"`
	ProcessMatch   []string               `yaml:"processMatch"`
	Repository     string                 `yaml:"repository"`
	ValidationNRQL string                 `yaml:"validationNrql"`
}

type variableConfig struct {
	Name    string `yaml:"name"`
	Prompt  string `yaml:"prompt"`
	Secret  bool   `secret:"prompt"`
	Default string `yaml:"default"`
}

type recipeInstallTarget struct {
	Type            string `yaml:"type"`
	OS              string `yaml:"os"`
	Platform        string `yaml:"platform"`
	PlatformFamily  string `yaml:"platformFamily"`
	PlatformVersion string `yaml:"platformVersion"`
	KernelVersion   string `yaml:"kernelVersion"`
	KernelArch      string `yaml:"kernelArch"`
}

type logMatch struct {
	Name       string             `yaml:"name"`
	File       string             `yaml:"file"`
	Attributes logMatchAttributes `yaml:"attributes,omitempty"`
	Pattern    string             `yaml:"pattern,omitempty"`
	Systemd    string             `yaml:"systemd,omitempty"`
}

type logMatchAttributes struct {
	LogType string `yaml:"logtype"`
}

type recipeFileFetcherImpl struct {
	HTTPGetFunc  func(string) (*http.Response, error)
	readFileFunc func(string) ([]byte, error)
}

func newRecipeFileFetcher() recipeFileFetcher {
	f := recipeFileFetcherImpl{}
	f.HTTPGetFunc = defaultHTTPGetFunc
	f.readFileFunc = defaultReadFileFunc
	return &f
}

func defaultHTTPGetFunc(recipeURL string) (*http.Response, error) {
	return http.Get(recipeURL)
}

func defaultReadFileFunc(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func (f *recipeFileFetcherImpl) fetchRecipeFile(recipeURL *url.URL) (*recipeFile, error) {
	response, err := f.HTTPGetFunc(recipeURL.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return toRecipeFile(string(body))
}

func (f *recipeFileFetcherImpl) loadRecipeFile(filename string) (*recipeFile, error) {
	out, err := f.readFileFunc(filename)
	if err != nil {
		return nil, err
	}
	return toRecipeFile(string(out))
}

func toRecipeFile(content string) (*recipeFile, error) {
	f, err := newRecipeFile(content)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func newRecipeFile(recipeFileString string) (*recipeFile, error) {
	var f recipeFile
	err := yaml.Unmarshal([]byte(recipeFileString), &f)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

func (f *recipeFile) String() (string, error) {
	out, err := yaml.Marshal(f)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (f *recipeFile) ToRecipe() (*recipe, error) {
	fileStr, err := f.String()
	if err != nil {
		return nil, err
	}

	r := recipe{
		File:           fileStr,
		Name:           f.Name,
		Description:    f.Description,
		Repository:     f.Repository,
		Keywords:       f.Keywords,
		ProcessMatch:   f.ProcessMatch,
		LogMatch:       f.LogMatch,
		ValidationNRQL: f.ValidationNRQL,
	}

	return &r, nil
}
