FROM node

ENV REPOS="https://github.com/rbell13/oa-sis"

RUN apt update && apt install -y openjdk-17-jdk subversion && rm -rf /var/lib/apt/lists/*

RUN npm install @openapitools/openapi-generator-cli -g

ADD ./pkg/pkg /usr/local/bin/pkg

ENTRYPOINT ["/usr/local/bin/pkg"]
EXPOSE 8080
