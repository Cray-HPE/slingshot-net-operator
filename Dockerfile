FROM arti.hpc.amslabs.hpecorp.net/baseos-docker-master-local/alpine:latest
COPY ./sshot-net-operator /root/sshot-net-operator
COPY ./entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
RUN chmod +x /root/sshot-net-operator
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]