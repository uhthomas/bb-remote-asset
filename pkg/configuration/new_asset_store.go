package configuration

import (
	"fmt"

	pb "github.com/buildbarn/bb-remote-asset/pkg/proto/configuration/bb_remote_asset"
	"github.com/buildbarn/bb-remote-asset/pkg/storage"
	asset_configuration "github.com/buildbarn/bb-remote-asset/pkg/storage/blobstore"
	"github.com/buildbarn/bb-storage/pkg/blobstore"
	blobstore_configuration "github.com/buildbarn/bb-storage/pkg/blobstore/configuration"
	"github.com/buildbarn/bb-storage/pkg/grpc"
)

// NewAssetStoreAndCASFromConfiguration creates an Asset Store and
// BlobAccess for the Content Addressable Storage.
func NewAssetStoreAndCASFromConfiguration(configuration *pb.AssetCacheConfiguration, grpcClientFactory grpc.ClientFactory, maximumMessageSizeBytes int) (storage.AssetStore, blobstore.BlobAccess, error) {
	switch backend := configuration.Backend.(type) {
	case *pb.AssetCacheConfiguration_BlobAccess:
		assetBlobAccessCreator := asset_configuration.NewAssetBlobAccessCreator(grpcClientFactory, maximumMessageSizeBytes)

		assetBlobAccess, err := blobstore_configuration.NewBlobAccessFromConfiguration(
			backend.BlobAccess.AssetStore,
			assetBlobAccessCreator)
		if err != nil {
			return nil, nil, err
		}
		assetStore := storage.NewBlobAccessAssetStore(assetBlobAccess.BlobAccess, maximumMessageSizeBytes)
		contentAddressableStorageInfo, err := blobstore_configuration.NewBlobAccessFromConfiguration(backend.BlobAccess.ContentAddressableStorage, blobstore_configuration.NewCASBlobAccessCreator(grpcClientFactory, maximumMessageSizeBytes))
		if err != nil {
			return nil, nil, err
		}
		contentAddressableStorage := contentAddressableStorageInfo.BlobAccess
		return assetStore, contentAddressableStorage, nil
	case *pb.AssetCacheConfiguration_ActionCache:
		contentAddressableStorage, actionCache, err := blobstore_configuration.NewCASAndACBlobAccessFromConfiguration(backend.ActionCache.Blobstore, grpcClientFactory, maximumMessageSizeBytes)
		if err != nil {
			return nil, nil, err
		}
		assetStore := storage.NewActionCacheAssetStore(actionCache, contentAddressableStorage, maximumMessageSizeBytes)
		return assetStore, contentAddressableStorage, nil
	}
	return nil, nil, fmt.Errorf("invalid backend configuration type: %T", configuration.Backend)
}
