FROM jameswoolfenden/ghat
WORKDIR /app
COPY . .
RUN yarn install --production
CMD ["node", "src/index.js"]
EXPOSE 3000
LABEL layer.0.author="James Woolfenden" layer.0.trace="202fe899-0eda-4da3-9e17-5d6feef8c46d" layer.0.tool="stevedore" git_repo="stevedore" git_org="JamesWoolfenden" git_file="examples/basic/Dockerfile"git_commit"37321c1fa74d62b2923a697c04c94910b9c210fc"
