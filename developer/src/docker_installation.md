# Docker Installation

This guide will cover how to install and setup docker on Ubuntu 24.04 LTS such that it can interact with a NVIDIA GPU and run multiple containers simultaneously.


## Prerequisites / Assumptions

This guide assumes that you are using [Linux Ubuntu 24.04 LTS](https://ubuntu.com/download/desktop/thank-you?version=24.04.2&architecture=amd64&lts=true) and don't have a previous docker installation on your system.


## Setting up Docker

Docker is a complex piece of virtualization software. It runs every piece of software in a seperate container, each defined by an image. That said, I'm not aware how one can integrate the NVIDIA Container Toolkit with the Docker Desktop version, which is why we will stick to the CLI version.

1) **Installing Docker**

    1) Clean machine | Uninstall all conflicting packages
        ```sh
        for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do sudo apt-get remove $pkg; done
        ``` 
    
    2) Setup Dockers repository

        ```sh
        # Add Docker's official GPG key:
        sudo apt-get update
        sudo apt-get install ca-certificates curl
        sudo install -m 0755 -d /etc/apt/keyrings
        sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
        sudo chmod a+r /etc/apt/keyrings/docker.asc

        # Add the repository to Apt sources:
        echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
        $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}") stable" | \
        sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt-get update
        ```

    3) Install docker packages

        ```sh
        sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
        ```
    
2) **Verfiying installation**
    ```sh
    # For the free world
    sudo docker run hello-world

    # For mainland china

    sudo docker run docker.1ms.run/hello-world
    ```

3) **_Optional set docker as start up service_**

    As of now docker is dorment requires manual restarting each time the terminal is closed using:

    ```sh
    sudo systemctl start docker
    ```

    If you want to have docker automatically available from the moment of boot up run:

    ```sh
    sudo systemctl enable docker.service
    sudo systemctl enable containerd.service
    ````

    To stop this behaviour, use `disable`:

    ```sh
    sudo systemctl disable docker.service
    sudo systemctl disable containerd.service
    ```

    > Note docker requires you to run it as a root-user. While it is possible to run it as a non-root-user, we will not cover that.


4) **Setting up Nvidia**

    Follow the nvidia setup guide, defined in `setup_nvidia_gpu.pdf`. It is recommended to complete this guide in a seperate terminal. 

5) **Finalising Installation**
    The following step is only necessary for chinese users to by pass the restrictions imposed by the great firewall of china.

    To by pass the great firewall of china legally, one has to use mirrors. I recommend [docker.1ms.run](https://docker.1ms.run) but I'm sure that there are plenty of other options.

    1) Open the `dockerd` configuration file
        ```sh
        sudo nano /etc/docker/daemon.json
        ```
    
    2) Overwrite the files content 

        ```sh
        
        ```

    3) Restart docker to apply changes

        ```sh
        sudo systemctl restart docker
        ```





## Sources

- [Installing Docker](https://docs.docker.com/engine/install/ubuntu/)
- [Configuring Docker after installation](https://docs.docker.com/engine/install/linux-postinstall/)



