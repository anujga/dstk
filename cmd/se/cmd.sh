 grpcurl -help
 grpcurl localhost:6001 list -plaintext
 grpcurl -plaintext localhost:6001 describe
 grpcurl -plaintext -d '{}' localhost:6001 dstk.SeClientApi/AllParts
 grpcurl -plaintext -d '{"workerId": 1}' localhost:6001 dstk.SeWorkerApi/MyParts
 grpcurl -plaintext -d '{"workerId": 2}' localhost:6001 dstk.SeWorkerApi/MyParts
