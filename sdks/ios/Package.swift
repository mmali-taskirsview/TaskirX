// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "AdxSDK",
    platforms: [
        .iOS(.v14)
    ],
    products: [
        .library(
            name: "AdxSDK",
            targets: ["AdxSDK"]
        )
    ],
    dependencies: [],
    targets: [
        .target(
            name: "AdxSDK",
            dependencies: [],
            path: "Sources/AdxSDK",
            resources: [
                .process("Resources")
            ]
        ),
        .testTarget(
            name: "AdxSDKTests",
            dependencies: ["AdxSDK"]
        )
    ]
)
