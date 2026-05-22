package com.rancy21.flowtask.model

enum class Priority(val label: String) {
    P1("P1"),
    P2("P2"),
    P3("P3");

    val order: Int get() = ordinal
}
