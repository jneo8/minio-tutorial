run:
	docker run -p 9000:9000 --name minio1 --rm\
	    -e "MINIO_ACCESS_KEY=user" \
	    -e "MINIO_SECRET_KEY=pwd" \
	    -v /tmp/minio-data:/data \
	    -v /tmp/minio-config:/root/.minio \
	    minio/minio server /data
