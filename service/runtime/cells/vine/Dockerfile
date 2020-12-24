FROM vine/vine

WORKDIR /

# This is just a hack to download most of the dependencies VINE and go VINE has
# We don't care it's not the latest version, just an approximation
RUN mkdir go && cd go && git clone https://github.com/VINE/VINE && \
    cd VINE && go install

COPY entrypoint.sh /
RUN chmod 755 entrypoint.sh
ENTRYPOINT ["bash", "/entrypoint.sh"]
