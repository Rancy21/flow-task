package com.rancy21.flowtask.ui.week

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rancy21.flowtask.data.entity.TaskEntity
import com.rancy21.flowtask.data.repository.TaskRepository
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import java.time.DayOfWeek
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.time.temporal.TemporalAdjusters

class WeekViewModel(
    private val taskRepository: TaskRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(WeekUiState())
    val uiState: StateFlow<WeekUiState> = _uiState.asStateFlow()

    private val today = LocalDate.now()
    private val weekStart = today.with(TemporalAdjusters.previousOrSame(DayOfWeek.MONDAY))
    private val weekEnd = today.with(TemporalAdjusters.nextOrSame(DayOfWeek.SUNDAY))
    private val formatter = DateTimeFormatter.ISO_LOCAL_DATE

    init {
        loadWeek()
    }

    fun loadWeek() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            taskRepository.getTasksForWeek(
                weekStart.format(formatter),
                weekEnd.format(formatter),
            ).collect { tasks ->
                _uiState.update { it.copy(tasks = tasks, isLoading = false) }
            }
        }
    }

    fun tasksByDay(): Map<LocalDate, List<TaskEntity>> {
        return _uiState.value.tasks.groupBy { task ->
            task.scheduledDate?.let { LocalDate.parse(it) } ?: today
        }.toSortedMap()
    }

    fun markDone(task: TaskEntity) {
        viewModelScope.launch {
            taskRepository.update(task.copy(status = "DONE", completedAt = java.time.Instant.now().toString()))
        }
    }

    fun delete(task: TaskEntity) {
        viewModelScope.launch {
            taskRepository.delete(task)
        }
    }
}
