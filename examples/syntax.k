type Car struct {
    doors : i32
    vendor : string
}

fn fib(n : i32) : i32 {
    if n < 3 {
        return 1
    }
    return fib(n-1) * fib(n-2)
}

fn main() : i32 {
    result := fib(1)

    if result == 1 {
        return 0
    }
}
