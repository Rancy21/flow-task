package com.rancy21.flowtask.data.dao

import androidx.room.Dao
import androidx.room.Delete
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Update
import com.rancy21.flowtask.data.entity.InboxEntity
import kotlinx.coroutines.flow.Flow

@Dao
interface InboxDao {
    @Query("SELECT * FROM inbox ORDER BY created_at DESC")
    fun getAll(): Flow<List<InboxEntity>>

    @Query("SELECT * FROM inbox WHERE id = :id")
    suspend fun getById(id: String): InboxEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insert(item: InboxEntity)

    @Update
    suspend fun update(item: InboxEntity)

    @Delete
    suspend fun delete(item: InboxEntity)

    @Query("DELETE FROM inbox WHERE id = :id")
    suspend fun deleteById(id: String)
}
