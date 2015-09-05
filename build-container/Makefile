# Change this to your DockerHub user when developing decap
DOCKER_USER := ae6rt

# No changes required below this line ------------

IMAGE := $(DOCKER_USER)/decap-build-base:latest

container:
	docker --tlsverify=false build -t $(IMAGE) .  

push: container
	docker --tlsverify=false push $(IMAGE)
