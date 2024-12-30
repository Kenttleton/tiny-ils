export interface Book {
    title: string,
    author: string,
    genre: Genre,
    borrowed: {
        date: Date,
        name: string,
    }
}

export interface Genre {

}