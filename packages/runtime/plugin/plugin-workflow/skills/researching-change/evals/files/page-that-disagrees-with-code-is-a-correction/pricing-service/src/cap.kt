fun dailyTotal(fares: List<Double>, cap: Double, discount: Double): Double {
    val capped = minOf(fares.sum(), cap)
    return capped * (1 - discount)
}
