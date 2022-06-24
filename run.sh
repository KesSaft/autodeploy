docker rm -f autodeploy
docker build . -t autodeploy
docker run --name=autodeploy --restart=always -p 127.0.0.1:10010:3000 -v /var/run/docker.sock:/var/run/docker.sock -v /projects:/projects autodeploy