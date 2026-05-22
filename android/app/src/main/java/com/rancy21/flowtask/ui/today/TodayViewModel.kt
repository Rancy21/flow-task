package com.rancy21.flowtask.ui.today

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rancy21.flowtask.data.entity.TaskEntity
import com.rancy21.flowtask.data.repository.TaskRepository
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.util.UUID

class TodayViewModel(
    private val taskRepository: TaskRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(TodayUiState())
    val uiState: StateFlow<TodayUiState> = _uiState.asStateFlow()

    private val todayDate: String = LocalDate.now().format(DateTimeFormatter.ISO_LOCAL_DATE)

    init {
        loadToday()
    }

    fun loadToday() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            taskRepository.getTasksForDate(todayDate).collect { tasks ->
                _uiState.update { it.copy(tasks = tasks, isLoading = false) }
            }
        }
    }

    fun markDone(task: TaskEntity) {
        viewModelScope.launch {
            val now = java.time.Instant.now().toString()
            taskRepository.update(
                task.copy(
                    status = "DONE",
                    completedAt = now,
                )
            )
        }
    }

    fun unSchedule(task: TaskEntity) {
        viewModelScope.launch {
            taskRepository.update(
                task.copy(
                    scheduledDate = null,
                    status = "INBOX",
                )
            )
        }
    }

    fun delete(task: TaskEntity) {
        viewModelScope.launch {
            taskRepository.delete(task)
        }
    }
}
