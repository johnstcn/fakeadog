FROM scratch
MAINTAINER Cian Johnston <public@cianjohnston.ie>
ADD fakeadog fakeadog
ENV HOST localhost
ENV PORT 8125
EXPOSE 8125
ENTRYPOINT ["/fakeadog"]
