package samples

// ComplexFunction demonstrates various complexity factors
func ComplexFunction(n int) int {
    // Initialize result
    result := 0
    
    // Iterate through numbers
    for i := 0; i < n; i++ {
        // Check if number is even or odd
        if i%2 == 0 {
            result += i    // Add even numbers
        } else {
            result -= i    // Subtract odd numbers
        }
    }
    
    return result
} 