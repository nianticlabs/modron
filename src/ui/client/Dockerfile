FROM node:18.11.0-buster-slim
RUN apt-get update && apt-get install -y gnupg wget python
RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'
RUN apt-get update && apt-get install -y google-chrome-stable

WORKDIR /usr/src/app
COPY . .
RUN chown -R node:node /usr/src/app
ENV CHROME_BIN=/usr/bin/google-chrome
ENV USER=node
ENV PATH="/home/node/.npm-global/bin:${PATH}"
ENV NPM_CONFIG_PREFIX="/home/node/.npm-global"
USER "${USER}"

RUN mkdir -p "${NPM_CONFIG_PREFIX}/lib"
RUN npm --global config set user "${USER}"
COPY package.json package-lock.json ./
RUN npm ci

ENTRYPOINT npm run test:ci
