FROM node:lts
COPY . /app
WORKDIR /app
RUN npm install
ENV NODE_OPTIONS='--loader ts-node/esm'
ENTRYPOINT ["node", "server.ts"]
