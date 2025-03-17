package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/git"

	"github.com/magefile/mage/sh"
)

const (
	imageBaseEnv          = "OCI_IMAGE_BASE"
	imageNameBaseFallback = "ocreg.invalid/coopnorge"
)

// Validate the content of a Dockerfile
func Validate(dockerfileContent string) error {
	dockerfilePath, cleanup, err := core.WriteTempFile("./var", "Dockerfile", dockerfileContent)
	if err != nil {
		return err
	}
	defer cleanup()

	dockerContext, cleanup, err := core.MkdirTemp()
	if err != nil {
		return err
	}
	defer cleanup()

	return sh.RunV("docker", "buildx", "build", "--check", "-f", dockerfilePath, dockerContext)
}

// BuildAndPush an OCI image for the provided platforms. Setting push to true
// will push the images to the registries. When push is true images are not
// tagged with latest.
func BuildAndPush(dockerfileContent, platforms, image, dockerContext, imagePath, metadatafile, app, binary string, push bool) error {
	versionTag := getVersionTag()
	versionTaggedImage := fmt.Sprintf("%s:%s", image, versionTag)
	latestImage := fmt.Sprintf("%s:latest", image)

	repoURL, err := git.RepoURL()
	if err != nil {
		return err
	}
	gitSHA256, err := git.SHA256()
	if err != nil {
		return err
	}

	err = createDirForOutput(imagePath)
	if err != nil {
		return err
	}
	err = createDirForOutput(metadatafile)
	if err != nil {
		return err
	}

	dockerfilePath, cleanup, err := core.WriteTempFile(core.OutputDir, "Dockerfile", dockerfileContent)
	if err != nil {
		return err
	}
	defer cleanup()

	args := []string{
		"buildx", "build",
		"--build-arg", fmt.Sprintf("GIT_REPOSITORY_URL=%s", repoURL),
		"--build-arg", fmt.Sprintf("GIT_COMMIT_SHA=%s", gitSHA256),
		"--build-arg", fmt.Sprintf("APP=%s", app),
		"--build-arg", fmt.Sprintf("BINARY=%s", binary),
		"--metadata-file", metadatafile,
		"--platform", platforms,
		"--output", fmt.Sprintf("type=image,push=%v", push),
		"--output", fmt.Sprintf("type=oci,dest=%s", imagePath),
		"-t", versionTaggedImage,
	}

	if !push {
		args = append(
			args,
			"-t", latestImage,
		)
	}

	args = append(args,
		"-f", dockerfilePath,
		dockerContext,
	)

	return sh.RunV("docker", args...)
}

// FindMetadataFiles ...
func FindMetadataFiles(base string) ([]string, error) {
	return filepath.Glob(fmt.Sprintf("%s/*/oci/*/metadata.json", base))
}

// Metadata ...
type Metadata struct {
	App       string
	Binary    string
	ImageName string
	Tag       string
}

// ParseMetadata ...
func ParseMetadata(filepath string) (Metadata, error) {
	if !path.IsAbs(filepath) {
		wd, err := os.Getwd()
		if err != nil {
			return Metadata{}, err
		}
		filepath = path.Join(wd, filepath)
	}
	file, err := os.Open(filepath)
	if err != nil {
		return Metadata{}, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return Metadata{}, err
	}

	var data map[string]any
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return Metadata{}, err
	}
	if len(data) == 0 {
		return Metadata{}, fmt.Errorf("no metadata found in: %s", filepath)
	}
	var imageNames []string
	for _, imageName := range strings.Split(fmt.Sprintf("%s", data["image.name"]), ",") {
		if strings.HasSuffix(imageName, ":latest") {
			continue
		}
		imageNames = append(imageNames, imageName)
	}
	if len(imageNames) != 1 {
		return Metadata{}, fmt.Errorf("image name not found in: %s", data["image.name"])
	}
	imageName := imageNames[0]

	return Metadata{
		ImageName: imageName,
		App:       getAppName(imageName),
		Binary:    getBinaryName(imageName),
		Tag:       getTag(imageName),
	}, nil
}

type binaryImage = map[string]string
type binaryImages = map[string]binaryImage

// AppImages ...
type AppImages = map[string]binaryImages

// Images ...
func Images(imageDir string) (AppImages, error) {
	metadataFiles, err := FindMetadataFiles(imageDir)
	if err != nil {
		return nil, err
	}

	result := make(AppImages)
	for _, file := range metadataFiles {
		metadata, err := ParseMetadata(file)
		if err != nil {
			return nil, err
		}
		if _, ok := result[metadata.App]; !ok {
			result[metadata.App] = make(binaryImages)
		}
		if _, ok := result[metadata.App][metadata.Binary]; !ok {
			result[metadata.App][metadata.Binary] = make(binaryImage)
		}
		result[metadata.App][metadata.Binary]["tag"] = metadata.Tag
		result[metadata.App][metadata.Binary]["image"] = metadata.ImageName
	}
	return result, nil
}

// FullyQualifiedlImageName ...
func FullyQualifiedlImageName(app, binary string) string {
	return fmt.Sprintf("%s/%s/%s", imageBase(), app, binary)
}

func imageBase() string {
	imageBase, ok := os.LookupEnv(imageBaseEnv)
	if !ok || imageBase == "" {
		imageBase = imageNameBaseFallback
	}
	return imageBase
}

func getAppName(imageName string) string {
	imageName = strings.TrimPrefix(imageName, fmt.Sprintf("%s/", imageBase()))
	imageName = strings.Split(imageName, ":")[0]
	return strings.Split(imageName, "/")[0]
}

func getBinaryName(imageName string) string {
	imageName = strings.TrimPrefix(imageName, fmt.Sprintf("%s/", imageBase()))
	imageName = strings.Split(imageName, ":")[0]
	return strings.Split(imageName, "/")[1]
}

func getTag(imageName string) string {
	return strings.Split(imageName, ":")[1]
}

func createDirForOutput(file string) error {
	dir := path.Dir(file)
	return os.MkdirAll(dir, 0700)
}

func getVersionTag() string {
	return time.Now().Format("v2006.01.02.150405")
}
