package com.rancy21.flowtask.data.dao

import androidx.room.Dao
import androidx.room.Delete
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Update
import com.rancy21.flowtask.data.entity.TaskEntity
import kotlinx.coroutines.flow.Flow

@Dao
interface TaskDao {
    @Query("SELECT * FROM tasks WHERE status != 'DONE' AND scheduled_date = :date ORDER BY priority ASC, created_at ASC")
    fun getTasksForDate(date: String): Flow<List<TaskEntity>>

    @Query("SELECT * FROM tasks WHERE status != 'DONE' AND scheduled_date >= :weekStart AND scheduled_date <= :weekEnd ORDER BY scheduled_date ASC, priority ASC, created_at ASC")
    fun getTasksForWeek(weekStart: String, weekEnd: String): Flow<List<TaskEntity>>

    @Query("SELECT * FROM tasks WHERE status = 'INBOX' ORDER BY priority ASC, created_at ASC")
    fun getInboxTasks(): Flow<List<TaskEntity>>

    @Query("SELECT * FROM tasks WHERE id = :id")
    suspend fun getById(id: String): TaskEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insert(task: TaskEntity)

    @Update
    suspend fun update(task: TaskEntity)

    @Delete
    suspend fun delete(task: TaskEntity)

    @Query("DELETE FROM tasks WHERE id = :id")
    suspend fun deleteById(id: String)
}
