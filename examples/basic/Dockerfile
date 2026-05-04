FROM jameswoolfenden/ghat:v0.1.23@sha256:163c3af710b5d1e112d9f268b63800c790622e70bf220e3459b029ce23ed9954
WORKDIR /app
COPY . .
RUN yarn install --production
CMD ["node", "src/index.js"]
EXPOSE 3000
