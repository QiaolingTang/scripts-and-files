# Install Operators From Bundle

```
operator-sdk run bundle $bundle-image
```

# Upgrade Operators
```
operator-sdk run bundle-upgrade $new-bundle-image
```

Notes:
1. if the catalogsource and the subscription are not in the same namespace, the upgrade will fail.
2. if the csv version in new bundle image is not changed, the upgrade will fail.


Ref: https://sdk.operatorframework.io/docs/olm-integration/tutorial-bundle/
