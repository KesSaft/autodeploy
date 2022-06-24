FROM golang:buster
WORKDIR /app
COPY . .
EXPOSE 3000
RUN go build -o autodeploy .
RUN chmod +x autodeploy
RUN curl -sSL https://get.docker.com/ | sh
RUN apt update && apt install -y curl git docker-compose
CMD ./autodeploy