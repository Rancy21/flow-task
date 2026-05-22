package com.rancy21.flowtask.data.database

import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import android.content.Context
import com.rancy21.flowtask.data.dao.InboxDao
import com.rancy21.flowtask.data.dao.NoteDao
import com.rancy21.flowtask.data.dao.TaskDao
import com.rancy21.flowtask.data.entity.InboxEntity
import com.rancy21.flowtask.data.entity.NoteEntity
import com.rancy21.flowtask.data.entity.TaskEntity

@Database(
    entities = [TaskEntity::class, NoteEntity::class, InboxEntity::class],
    version = 1,
    exportSchema = true,
)
abstract class FlowTaskDatabase : RoomDatabase() {
    abstract fun taskDao(): TaskDao
    abstract fun noteDao(): NoteDao
    abstract fun inboxDao(): InboxDao

    companion object {
        @Volatile
        private var INSTANCE: FlowTaskDatabase? = null

        fun getDatabase(context: Context): FlowTaskDatabase {
            return INSTANCE ?: synchronized(this) {
                val instance = Room.databaseBuilder(
                    context.applicationContext,
                    FlowTaskDatabase::class.java,
                    "flowtask.db",
                )
                    .build()
                INSTANCE = instance
                instance
            }
        }
    }
}
