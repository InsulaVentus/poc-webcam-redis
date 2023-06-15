.PHONY: redis
redis:
	@kubectl run redis --expose --image redis --port 6379

.PHONY: redis-stop
redis-stop:
	@kubectl delete service redis
	@kubectl delete pod redis

.PHONY: image-server
image-server:
	docker build --no-cache -t image-server:latest -f ./image-server/Dockerfile ./image-server
	docker tag image-server:latest k3d-registry.localhost:5000/image-server:latest
	docker push k3d-registry.localhost:5000/image-server:latest
	-kubectl delete -f ./image-server/deployment.yaml
	kubectl apply -f ./image-server/deployment.yaml

.PHONY: image-api
image-api:
	docker build --no-cache -t image-api:latest -f ./api/Dockerfile ./api
	docker tag image-api:latest k3d-registry.localhost:5000/image-api:latest
	docker push k3d-registry.localhost:5000/image-api:latest
	-kubectl delete -f ./api/deployment.yaml
	kubectl apply -f ./api/deployment.yaml

.PHONY: test
test:
	@hey -c 100 -z 10s http://localhost:8088/images?url=http%3A%2F%2Fimage-server%3A8888%2Fimage
