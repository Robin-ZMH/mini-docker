package src

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
)

func createContainerDirectories(containerID string) error {
	path := getPathOfContainer(containerID)
	dirs := []string{path + "/fs", path + "/fs/mnt", path + "/fs/upperdir", path + "/fs/workdir"}
	return createDirs(dirs)
}

func mountOverlayFileSystem(containerID, imageHash string) error {
	layers := getLayersOfImage(imageHash)
	containerFS := getFSHomeOfContainer(containerID)
	// mntOptions := "lowerdir=" + strings.Join(layers, ":") + ",upperdir=" + containerFS + "/upperdir,workdir=" + containerFS + "/workdir"
	mntOptions := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		strings.Join(layers, ":"),
		containerFS+"/upperdir",
		containerFS+"/workdir")
	return syscall.Mount("none", getMountPathOfContainer(containerID), "overlay", 0, mntOptions)
}

func unmountContainerFs(containerID string) error {
	mountPath := getMountPathOfContainer(containerID)
	return syscall.Unmount(mountPath, 0)
}

func clean_up(containerID string) {
	log.Println("Clean up...")
	_ = unmountContainerFs(containerID)
	_ = os.RemoveAll(getHomeOfContainers())
	log.Println("All temporary container files are removed")
}
