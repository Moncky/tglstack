package main

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"io/fs"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

// Return a slice of uniqe module direcotires
func tfRoot(files string) string {
	modulePaths := filepath.Dir(files)
	//modulePathMap := make([]string, 0, len())
	//for _, v := range {
	//	modulePathMap[v] = struct{}{}
	//}
	return modulePaths
}

// Get uniq module direcotires

func uniqModuleDirs(dirs []string) []string {
	moduleDirs := []string{}
	seen := make(map[string]bool)
	for _, val := range dirs {
		if _, in := seen[val]; !in {
			seen[val] = true
			moduleDirs = append(moduleDirs, val)
		}
	}
	return moduleDirs
}

// Get the abolute path the the users homedir
func userHome() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	homeDir := user.HomeDir
	return homeDir
}

// Returns a slice of string of terraform files that are not provider.tf
func tfFiles(rootDir string) []string {
	var tfFiles []string
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if info.IsDir() && info.Name() == ".git" || info.Name() == ".idea" {
			return filepath.SkipDir
		}
		// create the slice of filenames that are not provider.tf
		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			tfFiles = append(tfFiles, path)
			for i, v := range tfFiles {
				if v == "provider.tf" {
					tfFiles = append(tfFiles[:i], tfFiles[i+1:]...)
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return tfFiles
}

//Process config files, return a new config as a []byte
func processConfig(tfFile string) *hclwrite.File {
	src, err := ioutil.ReadFile(tfFile)
	//provider := filepath.Dir(tfFile) + "/providers.tf"
	hclfile, diags := hclwrite.ParseConfig(src, tfFile, hcl.Pos{Line: 1, Column: 1})

	var newConf = hclwrite.NewEmptyFile()
	rootBody := newConf.Body()

	if err != nil {
		panic(err)
	}
	if diags.HasErrors() {
		for _, diag := range diags {
			if diag.Subject != nil {
				fmt.Println("[%s:%d] %s: %s", diag.Subject.Filename, diag.Subject.Start.Line, diag.Summary, diag.Detail)
			} else {
				fmt.Printf("%s: %s", diag.Summary, diag.Detail)
			}
		}

	}
	// For each block of the type provider, remove that block
	for _, block := range hclfile.Body().Blocks() {
		if block.Type() == "terraform" {
			rootBody.AppendBlock(block)
			rootBody.AppendNewline()
			//hclfile.Body().RemoveBlock(block)
		}
		if block.Type() == "provider" {
			rootBody.AppendBlock(block)
			rootBody.AppendNewline()
			//hclfile.Body().RemoveBlock(block)
		}
	}

	newSrc := hclfile.Bytes()
	if bytes.Equal(newSrc, src) {
		fmt.Println("No Chages")
	}

	fmt.Printf("%s", newConf.Bytes())
	//fmt.Println("Writing to: " + provider)
	//err = ioutil.WriteFile(provider, newConf.Bytes(), 0644)
	err = ioutil.WriteFile(tfFile, newSrc, 0644)
	fmt.Println(err)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%s", newConf.Bytes())
	return newConf
}

func main() {
	// 1. detect all the direcories that have a terraformfile
	//  This is done with the tfFiles function that returns a slice of all tf files
	// 2. get a unice slice for each direcory
	//	a. tfRoot returns a string of the directory for any given tf file
	//  b. ** These need to be turned into a uniqe slice
	// 3. process the terraform files in each of those dirs
	// 4. write out the new provider.tf file

	//path := moduleRoot()
	modulesRoot := userHome() + "/sandbox/terraform/infrastructure-modules"
	//fmt.Println(tfRoot(modulesRoot))
	var allModuleDirs []string
	var moduleDirs []string
	// Get a uniq list of directoriesbthat contain at least one tf file
	for _, file := range tfFiles(modulesRoot) {
		allModuleDirs = append(allModuleDirs, tfRoot(file))
	}
	// Iterate the slice of unique module direcores and process each tf file, then create the new providers.tf file
	for _, module := range uniqModuleDirs(allModuleDirs) {
		moduleDirs = append(moduleDirs, module)
	}

	//Process the terraform files in each of the module directories
	for _, dir := range allModuleDirs {
		provider := dir + "/providers.tf"
		pcfg, err := os.Create(provider)
		if err != nil {
			panic(err)
		}
		for _, file := range tfFiles(dir) {
			fmt.Println("Processing: " + dir)
			fmt.Println("	File: " + file)
			newConfig := processConfig(file)
			fmt.Printf("%s", newConfig.Bytes())
			_, err := pcfg.Write(newConfig.Bytes())
			if err != nil {
				panic(err)
			}
		}

	}
}