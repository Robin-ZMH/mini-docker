package src

import (
	"encoding/json"
	"github.com/google/go-containerregistry/pkg/crane"
	"log"
	"os"
	"strings"
)

type tagToShaHex map[string]string
type imageDict map[string]tagToShaHex

type manifest []struct {
	Config   string
	RepoTags []string
	Layers   []string
}

type imageConfigDetails struct {
	Env []string `json:"Env"`
	Cmd []string `json:"Cmd"`
}

type imageConfig struct {
	Config imageConfigDetails `json:"config"`
}

// initImage will return the imageHash of the user input image,
// if that image not exist, download a new one and return the imageHash of the new image.
func initImage(input_img string) string {
	image, tag := getImageNameAndTag(input_img)
	if imageHash, ok := getimageHash(image, tag); ok {
		return imageHash
	} else {
		log.Printf("Downloading metadata for %s:%s, please wait...", image, tag)
		imageHash := downloadImage(strings.Join([]string{image, tag}, ":"))
		return imageHash
	}
}

// getImageNameAndTag split the input image string into image and tag,for example:
// python:3.8.10 -> [python, 3.8.10];
// python -> [python, latest].
func getImageNameAndTag(img string) (image, tag string) {
	s := strings.Split(img, ":")
	if len(s) > 1 {
		image, tag = s[0], s[1]
	} else {
		image = s[0]
		tag = "latest"
	}
	return
}

// getimageHash first load the json data into a imageDict,
// then check if the image and tag in the imageDict and
// return the hash code of the image if it exists.
func getimageHash(image, tag string) (string, bool) {
	imgDict := make(imageDict)
	parseImagesMetadata(&imgDict)
	if tagShaHex, ok := imgDict[image]; ok {
		shaHex, ok := tagShaHex[tag]
		return shaHex, ok
	}
	return "", false
}

// downloadImage pulls the image from remote and untar it to the images file,
// stores the meta information and clear the tar,
// and returns the hash code of the image.
func downloadImage(image string) string {
	// get the remote image information
	remoteImg, err := crane.Pull(image)
	if err != nil {
		log.Fatal(err)
	}
	manifest, err := remoteImg.Manifest()
	if err != nil {
		log.Fatal(err)
	}
	// get the firt 12 hash value of the image
	imageHash := manifest.Config.Digest.Hex
	log.Printf("ImageHash: %v\n", imageHash)

	// make a new dir and save the image as a tar file
	dir := getTarPath() + "/" + imageHash
	os.Mkdir(dir, 0755)
	path_tar := dir + "/package.tar"
	log.Printf("Downloading %s to %s\n", image, path_tar)
	must_ok(crane.SaveLegacy(remoteImg, image, path_tar))
	log.Printf("Successfully downloaded %s\n", image)

	// untar
	must_ok(untar(path_tar, dir))
	// decompress each layer of image
	processLayerTarballs(imageHash, dir)
	// update information of images.json
	storeImageMetadata(image, imageHash)
	// clear tar
	deleteTarFiles(imageHash)
	return imageHash
}

// processLayerTarballs read the information of manifest.json
// and extracts tar file of each layer into
// "dockerImagesPath/{image hash}/{layer hash}/fs".
func processLayerTarballs(imageHash, tarDir string) {
	manifestPath := tarDir + "/manifest.json"
	configPath := tarDir + "/" + imageHash + ".json"

	imgDir := getHomeOfImages() + "/" + imageHash
	log.Printf("Create image dir %s\n", imgDir)
	_ = os.Mkdir(imgDir, 0755)

	// untar the layer files, which will become the container root fs
	mani := readManifest(manifestPath)
	for _, layer := range mani[0].Layers {
		imageLayerDir := imgDir + "/" + layer[:12] + "/fs"
		_ = os.MkdirAll(imageLayerDir, 0755)
		src := tarDir + "/" + layer
		log.Printf("Uncompressing layer %s to: %s \n", src, imageLayerDir)
		untar(src, imageLayerDir)
	}
	log.Println("Uncompressing finish")

	must_ok(copyFile(manifestPath, getManifestOfImage(imageHash)))
	must_ok(copyFile(configPath, getConfigForImage(imageHash)))
}

// storeImageMetadata loads the data into imageDict
// and writes it into "dockerImagesPath/images.json".
func storeImageMetadata(src, imageHash string) {
	img, tag := getImageNameAndTag(src)
	imgDict := make(imageDict)
	shaHexDict := make(tagToShaHex)
	parseImagesMetadata(&imgDict)
	if imgDict[img] != nil {
		shaHexDict = imgDict[img]
	}
	shaHexDict[tag] = imageHash
	imgDict[img] = shaHexDict
	must_ok(marshalImageMetadata(imgDict))
}

func marshalImageMetadata(imgDict imageDict) error {
	data, err := json.Marshal(imgDict)
	if err != nil {
		return err
	}
	imgDictPath := getHomeOfImages() + "/" + "images.json"
	return os.WriteFile(imgDictPath, data, 0644)
}

// parseImagesMetadata load the data of json file that contains the images information
// into the imageDict,
// if that json file not exist, create a new one.
func parseImagesMetadata(imgDict *imageDict) {
	imgDictPath := getHomeOfImages() + "/" + "images.json"
	if _, err := os.Stat(imgDictPath); os.IsNotExist(err) {
		os.WriteFile(imgDictPath, []byte("{}"), 0644)
	}
	data, _ := os.ReadFile(imgDictPath)
	must_ok(json.Unmarshal(data, imgDict))
}
