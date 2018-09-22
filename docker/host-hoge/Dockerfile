FROM kinogmt/centos-ssh

RUN useradd hoge && \
    echo "hoge:testpass" | chpasswd && \
    mkdir -p /home/hoge/.ssh

ADD misc/testdata/sshr_keys/id_rsa.pub /home/hoge/.ssh

RUN touch /home/hoge/.ssh/authorized_keys && \
    chown hoge:hoge /home/hoge/.ssh/authorized_keys && \
    chmod 600 /home/hoge/.ssh/authorized_keys && \
    cat /home/hoge/.ssh/id_rsa.pub > /home/hoge/.ssh/authorized_keys