FROM alpine
COPY alita_robot /
ENTRYPOINT ["/alita_robot"]

LABEL org.opencontainers.image.authors="Divanshu Chauhan <divkix@divkix.me>"
LABEL org.opencontainers.image.url="https://divkix.me"
LABEL org.opencontainers.image.source="https://github.com/divideprojects/Alita_Robot"
LABEL org.opencontainers.image.title="Alita Go Robot"
LABEL org.opencontainers.image.description="Official Alita Go Robot Docker Image"
LABEL org.opencontainers.image.vendor="Divkix"
