##
# This is place inside `/etc/profile.d/99-firehose-acme.sh`
# on built system an executed to provide message to use when they
# connect on the box.
export PATH=$PATH:/app

# If we are in a "node-manager" image, display special scripts motd#
#
# *Note* Our (i.e. firehose-acme) Mindreader data directory is at the root
#        `/data` mount point. Inside it, the `dummmy-dm-indexer` binary
#        itself creates a `data` subfolder. This is why we have `/data/data`
#        here as the path to check if we are inside a Node Manager instance
#        or a pure process (i.e. without Node Manager)
if [[ -d /data/data  ]]; then
    cat /etc/motd_node_manager
else
    cat /etc/motd_generic
fi
