package com.taskirx.example

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.taskirx.sdk.AdxSDK
import com.taskirx.sdk.ads.*
import com.taskirx.sdk.config.AdSize
import com.taskirx.sdk.config.AdxConfig
import kotlinx.coroutines.launch

/**
 * Example Android app demonstrating AdxSDK usage
 */
class MainActivity : ComponentActivity() {
    
    private lateinit var interstitialAd: AdxInterstitial
    private lateinit var rewardedVideoAd: AdxRewardedVideo
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Initialize SDK (do this once in Application class)
        AdxSDK.init(
            context = this,
            publisherId = "your-publisher-id",
            config = AdxConfig(
                apiEndpoint = "http://10.0.2.2:3000", // For Android emulator
                enableDebug = true
            )
        )
        
        // Create interstitial ad
        interstitialAd = AdxInterstitial(this)
        
        // Create rewarded video ad
        rewardedVideoAd = AdxRewardedVideo(this)
        
        setContent {
            MaterialTheme {
                MainScreen(
                    interstitialAd = interstitialAd,
                    rewardedVideoAd = rewardedVideoAd
                )
            }
        }
    }
    
    override fun onDestroy() {
        super.onDestroy()
        interstitialAd.destroy()
        rewardedVideoAd.destroy()
    }
}

@Composable
fun MainScreen(
    interstitialAd: AdxInterstitial,
    rewardedVideoAd: AdxRewardedVideo
) {
    val scope = rememberCoroutineScope()
    val snackbarHostState = remember { SnackbarHostState() }
    var userCoins by remember { mutableStateOf(0) }
    
    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
            TopAppBar(
                title = { Text("AdxSDK Demo") },
                actions = {
                    Text("Coins: $userCoins", modifier = Modifier.padding(horizontal = 16.dp))
                }
            )
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Section 1: Banner Ads
            item {
                Text("Banner Ad (320x50)", style = MaterialTheme.typography.titleMedium)
                AdxBannerView(
                    placementId = "banner-home",
                    adSize = AdSize.BANNER_320x50,
                    onAdLoaded = {
                        scope.launch {
                            snackbarHostState.showSnackbar("Banner ad loaded")
                        }
                    },
                    onAdFailed = { error ->
                        scope.launch {
                            snackbarHostState.showSnackbar("Banner failed: ${error.message}")
                        }
                    },
                    onAdClicked = {
                        scope.launch {
                            snackbarHostState.showSnackbar("Banner clicked!")
                        }
                    }
                )
            }
            
            item {
                Divider()
            }
            
            // Section 2: Interstitial Ad
            item {
                Text("Interstitial Ad", style = MaterialTheme.typography.titleMedium)
                Text("Full-screen ad that appears after delay", style = MaterialTheme.typography.bodySmall)
                
                Spacer(modifier = Modifier.height(8.dp))
                
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Button(
                        onClick = {
                            interstitialAd.load(
                                placementId = "interstitial-level-complete",
                                onAdLoaded = {
                                    scope.launch {
                                        snackbarHostState.showSnackbar("Interstitial loaded")
                                    }
                                },
                                onAdFailed = { error ->
                                    scope.launch {
                                        snackbarHostState.showSnackbar("Failed: ${error.message}")
                                    }
                                }
                            )
                        }
                    ) {
                        Text("Load Interstitial")
                    }
                    
                    Button(
                        onClick = {
                            if (interstitialAd.isReady()) {
                                interstitialAd.show(
                                    onAdShown = {
                                        scope.launch {
                                            snackbarHostState.showSnackbar("Interstitial shown")
                                        }
                                    },
                                    onAdClosed = {
                                        scope.launch {
                                            snackbarHostState.showSnackbar("Interstitial closed")
                                        }
                                    }
                                )
                            } else {
                                scope.launch {
                                    snackbarHostState.showSnackbar("Ad not ready. Load first!")
                                }
                            }
                        },
                        enabled = interstitialAd.isReady()
                    ) {
                        Text("Show Interstitial")
                    }
                }
            }
            
            item {
                Divider()
            }
            
            // Section 3: Rewarded Video
            item {
                Text("Rewarded Video Ad", style = MaterialTheme.typography.titleMedium)
                Text("Watch video to earn 10 coins", style = MaterialTheme.typography.bodySmall)
                
                Spacer(modifier = Modifier.height(8.dp))
                
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Button(
                        onClick = {
                            rewardedVideoAd.load(
                                placementId = "rewarded-extra-lives",
                                onAdLoaded = {
                                    scope.launch {
                                        snackbarHostState.showSnackbar("Rewarded video loaded")
                                    }
                                },
                                onAdFailed = { error ->
                                    scope.launch {
                                        snackbarHostState.showSnackbar("Failed: ${error.message}")
                                    }
                                }
                            )
                        }
                    ) {
                        Text("Load Rewarded")
                    }
                    
                    Button(
                        onClick = {
                            if (rewardedVideoAd.isReady()) {
                                rewardedVideoAd.show(
                                    onAdShown = {
                                        scope.launch {
                                            snackbarHostState.showSnackbar("Rewarded video shown")
                                        }
                                    },
                                    onRewarded = { reward ->
                                        userCoins += reward.amount
                                        scope.launch {
                                            snackbarHostState.showSnackbar("Earned ${reward.amount} coins!")
                                        }
                                    },
                                    onAdClosed = {
                                        scope.launch {
                                            snackbarHostState.showSnackbar("Rewarded video closed")
                                        }
                                    }
                                )
                            } else {
                                scope.launch {
                                    snackbarHostState.showSnackbar("Ad not ready. Load first!")
                                }
                            }
                        },
                        enabled = rewardedVideoAd.isReady()
                    ) {
                        Text("Show Rewarded")
                    }
                }
            }
            
            item {
                Divider()
            }
            
            // Section 4: Native Ad
            item {
                Text("Native Ad", style = MaterialTheme.typography.titleMedium)
                Text("Customizable ad that matches your app design", style = MaterialTheme.typography.bodySmall)
                
                Spacer(modifier = Modifier.height(8.dp))
                
                // TODO: Implement native ad loading and display
                Card {
                    Text(
                        text = "Native ad will appear here",
                        modifier = Modifier.padding(16.dp)
                    )
                }
            }
            
            item {
                Divider()
            }
            
            // Section 5: SDK Info
            item {
                Text("SDK Information", style = MaterialTheme.typography.titleMedium)
                Card {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text("SDK Version: ${AdxSDK.getVersion()}")
                        Text("Publisher ID: ${AdxSDK.publisherId}")
                        Text("User ID: ${AdxSDK.getUserId()}")
                        Text("Session ID: ${AdxSDK.sessionId}")
                        
                        Spacer(modifier = Modifier.height(8.dp))
                        
                        Button(
                            onClick = {
                                scope.launch {
                                    val gaid = AdxSDK.getAdvertisingId()
                                    snackbarHostState.showSnackbar(
                                        "GAID: ${gaid ?: "Not available"}"
                                    )
                                }
                            }
                        ) {
                            Text("Get Advertising ID")
                        }
                    }
                }
            }
        }
    }
}
