IMAGE=ae6rt/decap-build-base:latest

container:
	docker --tlsverify=false build -t $(IMAGE) .  

push: container
	docker --tlsverify=false push $(IMAGE)
