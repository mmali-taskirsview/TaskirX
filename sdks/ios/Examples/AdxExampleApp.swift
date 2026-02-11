import SwiftUI

@main
struct AdxExampleApp: App {
    
    init() {
        // Initialize AdxSDK
        AdxSDK.initialize(
            publisherId: "your-publisher-id",
            configuration: AdxConfiguration(
                apiEndpoint: "http://localhost:3000",
                enableDebug: true,
                testMode: true
            )
        )
    }
    
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
