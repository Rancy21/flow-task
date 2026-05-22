package com.rancy21.flowtask

import android.app.Application
import com.rancy21.flowtask.di.appModule
import org.koin.android.ext.koin.androidContext
import org.koin.core.context.startKoin

class FlowTaskApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        startKoin {
            androidContext(this@FlowTaskApplication)
            modules(appModule)
        }
    }
}
