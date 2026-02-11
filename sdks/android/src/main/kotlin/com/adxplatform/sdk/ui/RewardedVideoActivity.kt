package com.taskirx.sdk.ui

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.media3.common.MediaItem
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.PlayerView
import com.taskirx.sdk.AdxSDK
import com.taskirx.sdk.ads.AdxReward
import com.taskirx.sdk.ads.AdxRewardedVideo
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/**
 * Full-screen rewarded video ad activity
 */
class RewardedVideoActivity : ComponentActivity() {
    
    private var bidId: String? = null
    private var videoUrl: String? = null
    private var clickUrl: String? = null
    private var exoPlayer: ExoPlayer? = null
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        bidId = intent.getStringExtra("bidId")
        videoUrl = intent.getStringExtra("videoUrl")
        clickUrl = intent.getStringExtra("clickUrl")
        
        setContent {
            MaterialTheme {
                RewardedVideoAdScreen(
                    videoUrl = videoUrl,
                    onVideoCompleted = {
                        // Give reward
                        val reward = AdxReward(type = "coins", amount = 10)
                        AdxRewardedVideo.rewardedVideoCallbacks?.onRewarded?.invoke(reward)
                        
                        // Track conversion
                        kotlinx.coroutines.GlobalScope.launch {
                            bidId?.let { 
                                AdxSDK.getNetworkClient().trackConversion(it, "video_complete")
                            }
                        }
                    },
                    onClose = {
                        AdxRewardedVideo.rewardedVideoCallbacks?.onAdClosed?.invoke()
                        finish()
                    },
                    onAdClick = {
                        // Track click
                        kotlinx.coroutines.GlobalScope.launch {
                            bidId?.let { AdxSDK.getNetworkClient().trackClick(it) }
                        }
                        
                        // Open URL
                        clickUrl?.let { url ->
                            try {
                                val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
                                startActivity(intent)
                            } catch (e: Exception) {
                                AdxSDK.logError("Failed to open URL", e)
                            }
                        }
                    },
                    onPlayerReady = { player ->
                        exoPlayer = player
                    }
                )
            }
        }
    }
    
    override fun onDestroy() {
        super.onDestroy()
        exoPlayer?.release()
    }
}

@Composable
private fun RewardedVideoAdScreen(
    videoUrl: String?,
    onVideoCompleted: () -> Unit,
    onClose: () -> Unit,
    onAdClick: () -> Unit,
    onPlayerReady: (ExoPlayer) -> Unit
) {
    val context = androidx.compose.ui.platform.LocalContext.current
    var showCloseButton by remember { mutableStateOf(false) }
    var videoCompleted by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()
    
    Box(
        modifier = Modifier.fillMaxSize()
    ) {
        // Video player
        videoUrl?.let { url ->
            AndroidView(
                factory = { ctx ->
                    PlayerView(ctx).apply {
                        val player = ExoPlayer.Builder(ctx).build()
                        this.player = player
                        
                        val mediaItem = MediaItem.fromUri(url)
                        player.setMediaItem(mediaItem)
                        player.prepare()
                        player.playWhenReady = true
                        
                        // Listen for playback completion
                        player.addListener(object : androidx.media3.common.Player.Listener {
                            override fun onPlaybackStateChanged(state: Int) {
                                if (state == androidx.media3.common.Player.STATE_ENDED) {
                                    if (!videoCompleted) {
                                        videoCompleted = true
                                        onVideoCompleted()
                                        showCloseButton = true
                                    }
                                }
                            }
                        })
                        
                        onPlayerReady(player)
                    }
                },
                modifier = Modifier.fillMaxSize()
            )
        }
        
        // Skip button (after 5 seconds)
        LaunchedEffect(Unit) {
            delay(5000)
            if (!videoCompleted) {
                showCloseButton = true
            }
        }
        
        // Close/Skip button
        if (showCloseButton) {
            Column(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(16.dp)
            ) {
                if (videoCompleted) {
                    Button(onClick = onClose) {
                        Text("Close")
                    }
                } else {
                    TextButton(onClick = onClose) {
                        Text("Skip (no reward)")
                    }
                }
            }
        }
        
        // Reward indicator
        if (videoCompleted) {
            Card(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(16.dp)
            ) {
                Text(
                    text = "🎉 Reward earned!",
                    modifier = Modifier.padding(16.dp),
                    style = MaterialTheme.typography.titleMedium
                )
            }
        }
    }
}
