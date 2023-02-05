# mini-docker
a mini-docker that can manage images(download from docker hub), create containers(use linux namespaces to isolate the containers), execute processes in existing containers like docker.  
mini-docker utilizes the Overlay file system when creating new containers.  
# how to run
sudo ./mini-docker run [image:tag] [cmds]

