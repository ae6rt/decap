DOCKER_USER := ae6rt

# No changes required below this line ------------

IMAGE := $(DOCKER_USER)/java7:latest

container:
	docker --tlsverify=false build -t $(IMAGE) .  

push: container
	docker --tlsverify=false push $(IMAGE)
