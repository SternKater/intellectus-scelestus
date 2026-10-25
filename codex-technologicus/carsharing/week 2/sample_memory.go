package sample

import (
	"fmt"
)

func getFirstThreeElements() []int {
    hugeData := make([]int, 1_000_000) // 8mb
    for i := range hugeData {
        hugeData[i] = i
    }
    return hugeData[:3] 
}


//=============================

type User struct {
    ID   int
    Name string
}

func createUser() *User {
    u := User{ID: 1, Name: "Alice"}
    return &u
}


//===============================

func main() {
    sidebar := []int{10, 20, 30}
    var result []*int

    for _, v := range sidebar {
        result = append(result, &v)
    }
    // ?
    for _, ptr := range result {
        fmt.Print(*ptr, " ")
    }
}

//===================================

func main() {
    m := make(map[int][1024]byte) // Тяжелые значения
    for i := 0; i < 100_000; i++ {
        m[i] = [1024]byte{}
    }
    
    // delete all
    for k := range m {
        delete(m, k)
    }
    // free?
}


//=========================================

func BenchmarkContext(b *testing.B) {
    var x int = 42
    for i := 0; i < b.N; i++ {
        // Какое скрытое действие с памятью происходит здесь?
        takeInterface(x) 
    }
}

func takeInterface(i interface{}) {}
