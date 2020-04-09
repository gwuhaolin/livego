FROM node:13.12

ENV WEBUI_DIR /src/webui/
ENV STATIC_DIR /src/static/
RUN mkdir -p $WEBUI_DIR
RUN mkdir -p $STATIC_DIR

COPY package.json $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN npm install

COPY . $WEBUI_DIR/
