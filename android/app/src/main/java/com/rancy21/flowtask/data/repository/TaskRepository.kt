package com.rancy21.flowtask.data.repository

import com.rancy21.flowtask.data.dao.TaskDao
import com.rancy21.flowtask.data.entity.TaskEntity
import kotlinx.coroutines.flow.Flow

class TaskRepository(private val taskDao: TaskDao) {
    fun getTasksForDate(date: String): Flow<List<TaskEntity>> = taskDao.getTasksForDate(date)

    fun getTasksForWeek(weekStart: String, weekEnd: String): Flow<List<TaskEntity>> =
        taskDao.getTasksForWeek(weekStart, weekEnd)

    fun getInboxTasks(): Flow<List<TaskEntity>> = taskDao.getInboxTasks()

    suspend fun getById(id: String): TaskEntity? = taskDao.getById(id)

    suspend fun save(task: TaskEntity) = taskDao.insert(task)

    suspend fun update(task: TaskEntity) = taskDao.update(task)

    suspend fun delete(task: TaskEntity) = taskDao.delete(task)

    suspend fun deleteById(id: String) = taskDao.deleteById(id)

    suspend fun upsert(task: TaskEntity) {
        if (taskDao.getById(task.id) != null) {
            taskDao.update(task)
        } else {
            taskDao.insert(task)
        }
    }
}
