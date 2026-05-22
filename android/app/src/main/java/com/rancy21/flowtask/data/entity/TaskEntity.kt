package com.rancy21.flowtask.data.entity

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "tasks")
data class TaskEntity(
    @PrimaryKey val id: String,
    val title: String,
    val description: String?,
    val priority: String,
    val status: String,
    @ColumnInfo(name = "scheduled_date") val scheduledDate: String?,
    @ColumnInfo(name = "created_at") val createdAt: String,
    @ColumnInfo(name = "completed_at") val completedAt: String?,
    @ColumnInfo(name = "updated_at") val updatedAt: String = "",
)
