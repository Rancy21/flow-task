package com.rancy21.flowtask

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.rancy21.flowtask.ui.navigation.FlowTaskApp
import com.rancy21.flowtask.ui.theme.FlowTaskTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            FlowTaskTheme {
                FlowTaskApp()
            }
        }
    }
}
