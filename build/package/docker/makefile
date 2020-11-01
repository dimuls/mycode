.PHONY:
	all
	mycode-run
	mycode-run-docker
	mycode-c-docker
	mycode-cpp-docker
	mycode-go-docker
	mycode-java-docker
	mycode-pascal-docker
	mycode-python-docker

all: mycode-c-docker mycode-cpp-docker mycode-go-docker mycode-java-docker \
	mycode-pascal-docker mycode-python-docker clean

clean:
	rm -f mycode-run

mycode-run:
	CGO_ENABLED=0 go build -o mycode-run ../../../cmd/mycode-run

mycode-run-docker: mycode-run
	docker build -t mycode-run -f mycode-run.dockerfile .

mycode-c-docker: mycode-run-docker
	docker build -t mycode-c -f mycode-c.dockerfile .

mycode-cpp-docker: mycode-run-docker
	docker build -t mycode-cpp -f mycode-cpp.dockerfile .

mycode-go-docker: mycode-run-docker
	docker build -t mycode-go -f mycode-go.dockerfile .

mycode-java-docker: mycode-run-docker
	docker build -t mycode-java -f mycode-java.dockerfile .

mycode-pascal-docker: mycode-run-docker
	docker build -t mycode-pascal -f mycode-pascal.dockerfile .

mycode-python-docker: mycode-run-docker
	docker build -t mycode-python -f mycode-python.dockerfile .
