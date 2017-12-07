FROM scratch
MAINTAINER Cian Johnston <public@cianjohnston.ie>

ADD fakeadog fakeadog
ENV HOST 0.0.0.0
ENV PORT 8125

EXPOSE 8125/udp

ENTRYPOINT ["/fakeadog"]
