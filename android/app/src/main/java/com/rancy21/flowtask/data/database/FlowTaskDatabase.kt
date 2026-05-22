package com.rancy21.flowtask.data.database

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.migration.Migration
import androidx.sqlite.db.SupportSQLiteDatabase
import com.rancy21.flowtask.data.dao.InboxDao
import com.rancy21.flowtask.data.dao.NoteDao
import com.rancy21.flowtask.data.dao.TaskDao
import com.rancy21.flowtask.data.entity.InboxEntity
import com.rancy21.flowtask.data.entity.NoteEntity
import com.rancy21.flowtask.data.entity.TaskEntity

@Database(
    entities = [TaskEntity::class, NoteEntity::class, InboxEntity::class],
    version = 2,
    exportSchema = false,
)
abstract class FlowTaskDatabase : RoomDatabase() {
    abstract fun taskDao(): TaskDao
    abstract fun noteDao(): NoteDao
    abstract fun inboxDao(): InboxDao

    companion object {
        @Volatile
        private var INSTANCE: FlowTaskDatabase? = null

        private val MIGRATION_1_2 = object : Migration(1, 2) {
            override fun migrate(db: SupportSQLiteDatabase) {
                db.execSQL("ALTER TABLE tasks ADD COLUMN updated_at TEXT NOT NULL DEFAULT ''")
                db.execSQL("ALTER TABLE notes ADD COLUMN updated_at TEXT NOT NULL DEFAULT ''")
                db.execSQL("ALTER TABLE inbox ADD COLUMN updated_at TEXT NOT NULL DEFAULT ''")
            }
        }

        fun getDatabase(context: Context): FlowTaskDatabase {
            return INSTANCE ?: synchronized(this) {
                val instance = Room.databaseBuilder(
                    context.applicationContext,
                    FlowTaskDatabase::class.java,
                    "flowtask.db",
                )
                    .addMigrations(MIGRATION_1_2)
                    .build()
                INSTANCE = instance
                instance
            }
        }
    }
}
