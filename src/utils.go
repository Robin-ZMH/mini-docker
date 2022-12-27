package src

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)


const dockerHomePath = "/tmp/miniDocker"
const dockerTarPath = dockerHomePath + "/tar"
const dockerImagesPath = dockerHomePath + "/images"
const dockerContainersPath = dockerHomePath + "/containers"

func must_ok(err error) {
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func InitDockerDirs() {
	log.Println("initialize Docker Dirs...")
	dirs := []string{dockerHomePath, dockerTarPath, dockerImagesPath, dockerContainersPath}
	must_ok(createDirs(dirs))
}

func createDirs(dirs []string) error {
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func genRandID() string {
	randBytes := make([]byte, 6)
	rand.Read(randBytes)
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x",
		randBytes[0], randBytes[1], randBytes[2],
		randBytes[3], randBytes[4], randBytes[5])
}

func untar(tarball, target string) error {
	hardLinks := make(map[string]string)
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()

		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue

		case tar.TypeLink:
			/* Store details of hard links, which we process finally */
			linkPath := filepath.Join(target, header.Linkname)
			linkPath2 := filepath.Join(target, header.Name)
			hardLinks[linkPath2] = linkPath
			continue

		case tar.TypeSymlink:
			linkPath := filepath.Join(target, header.Name)
			if err := os.Symlink(header.Linkname, linkPath); err != nil {
				if os.IsExist(err) {
					continue
				}
				return err
			}
			continue

		case tar.TypeReg:
			/* Ensure any missing directories are created */
			if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
				os.MkdirAll(filepath.Dir(path), 0755)
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if os.IsExist(err) {
				continue
			}
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tarReader)
			file.Close()
			if err != nil {
				return err
			}

		default:
			log.Printf("Warning: File type %d unhandled by untar function!\n", header.Typeflag)
		}
	}

	/* To create hard links the targets must exist, so we do this finally */
	for k, v := range hardLinks {
		if err := os.Link(v, k); err != nil {
			return err
		}
	}
	return nil
}

func deleteTarFiles(imageHash string) {
	path := getTarPath() + "/" + imageHash
	must_ok(os.RemoveAll(path))
}

func readManifest(manifestPath string) (mani manifest) {
	data, err := os.ReadFile(manifestPath)

	if err != nil {
		log.Fatalf("Could not read %s.\n", manifestPath)
	}

	if err := json.Unmarshal(data, &mani); err != nil {
		log.Fatalf("Could not Unmarshal %s.\n", string(data))
	}

	if len(mani) == 0 || len(mani[0].Layers) == 0 || len(mani) > 1 {
		log.Fatalf("Could not handle %s.\n", manifestPath)
	}

	return 
}

func readContainerConfig(imageHash string) (imgconf imageConfig) {
	configPath := getConfigForImage(imageHash)
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln("Could not read image config file")
	}
	
	must_ok(json.Unmarshal(data, &imgconf))
	return
}

func getHomeOfImages() string {
	return dockerImagesPath
}

func getPathOfImage(imageHash string) string {
	return getHomeOfImages() + "/" + imageHash
}

func getTarPath() string {
	return dockerTarPath
}

func getHomeOfContainers() string {
	return dockerContainersPath
}

func getPathOfContainer(containerID string) string {
	return getHomeOfContainers() + "/" + containerID
}

func getManifestOfImage(imageHash string) string {
	return getPathOfImage(imageHash) + "/manifest.json"
}

func getConfigForImage(imageHash string) string {
	return getPathOfImage(imageHash) + "/" + imageHash + ".json"
}

func getFSHomeOfContainer(contanerID string) string {
	return getHomeOfContainers() + "/" + contanerID + "/fs"
}

func getMountPathOfContainer(contanerID string) string {
	return getFSHomeOfContainer(contanerID) + "/mnt"
}

func getLayersOfImage(imageHash string) (layers []string) {
	manifestPath := getManifestOfImage(imageHash)
	mani := readManifest(manifestPath)
	imagePath := getPathOfImage(imageHash)
	for _, layer := range mani[0].Layers {
		// must append like this, because upper(new) layers must override lower(low) layers
		layers = append([]string{imagePath + "/" + layer[:12] + "/fs"}, layers...)
		/*
		if use append like:
		layers = append(layers, imagePath + "/" + layer[:12] + "/fs"),
		the old layers will cover the new layers.
		*/
	}
	return 
}
