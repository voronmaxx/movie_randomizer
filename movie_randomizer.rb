# movie_randomizer.rb
require 'json'
require 'date'
require 'csv'

class Movie
  attr_accessor :id, :title, :genre, :director, :year, :watched, :rating, :notes, :added_date

  def initialize(id, title, genre, director, year, watched, rating, notes = "", added_date = Date.today.to_s)
    @id = id
    @title = title
    @genre = genre
    @director = director
    @year = year
    @watched = watched
    @rating = rating
    @notes = notes
    @added_date = added_date
  end

  def to_h
    { id: @id, title: @title, genre: @genre, director: @director, year: @year,
      watched: @watched, rating: @rating, notes: @notes, added_date: @added_date }
  end

  def self.from_h(hash)
    Movie.new(hash[:id], hash[:title], hash[:genre], hash[:director], hash[:year],
              hash[:watched], hash[:rating], hash[:notes], hash[:added_date])
  end
end

class MovieRandomizer
  attr_reader :movies

  def initialize
    @movies = []
    @next_id = 1
  end

  def add_movie(title, genre, director, year, watched, rating, notes = "")
    raise "Оценка должна быть от 1 до 10" unless (1..10).include?(rating)
    raise "Год должен быть от 1900 до #{Date.today.year}" unless (1900..Date.today.year).include?(year)
    raise "Название и жанр не могут быть пустыми" if title.empty? || genre.empty?
    director = "Неизвестен" if director.empty?
    movie = Movie.new(@next_id, title, genre, director, year, watched, rating, notes)
    @movies << movie
    @next_id += 1
    movie
  end

  def find_movie(id)
    @movies.find { |m| m.id == id }
  end

  def edit_movie(id, **kwargs)
    movie = find_movie(id)
    return false unless movie
    kwargs.each do |key, value|
      movie.send("#{key}=", value) if movie.respond_to?("#{key}=")
    end
    true
  end

  def delete_movie(id)
    movie = find_movie(id)
    return false unless movie
    @movies.delete(movie)
    true
  end

  def unwatched
    @movies.select { |m| !m.watched }
  end

  def by_genre(genre)
    @movies.select { |m| m.genre.downcase == genre.downcase }
  end

  def random_movie(genre = nil)
    pool = genre ? by_genre(genre) : unwatched
    pool = @movies if pool.empty?  # если нет непросмотренных или по жанру, берём все
    pool = by_genre(genre) if genre && pool.empty?  # если по жанру нет, пробуем все по жанру
    pool.empty? ? nil : pool.sample
  end

  def stats
    total = @movies.size
    watched_count = @movies.count(&:watched)
    unwatched = total - watched_count
    ratings = @movies.select(&:watched).map(&:rating)
    avg_rating = ratings.empty? ? 0 : ratings.sum.to_f / ratings.size
    genres = Hash.new(0)
    @movies.each { |m| genres[m.genre] += 1 }
    { total: total, watched: watched_count, unwatched: unwatched, avg_rating: avg_rating, genres: genres }
  end

  def save_to_file(filename = "movies_data.json")
    data = { movies: @movies.map(&:to_h) }
    File.write(filename, JSON.pretty_generate(data))
  end

  def load_from_file(filename = "movies_data.json")
    return unless File.exist?(filename)
    data = JSON.parse(File.read(filename), symbolize_names: true)
    @movies.clear
    data[:movies].each do |item|
      movie = Movie.from_h(item)
      @movies << movie
      @next_id = movie.id + 1 if movie.id >= @next_id
    end
  rescue JSON::ParserError
    puts "Ошибка чтения файла."
  end

  def export_csv(filename = "movies_export.csv")
    CSV.open(filename, "w", col_sep: ";") do |csv|
      csv << ["ID", "Название", "Жанр", "Режиссёр", "Год", "Просмотрен", "Оценка", "Заметки", "Дата добавления"]
      @movies.each do |m|
        csv << [m.id, m.title, m.genre, m.director, m.year, m.watched ? "Да" : "Нет", m.rating, m.notes, m.added_date]
      end
    end
  end

  def import_csv(filename = "movies_export.csv")
    unless File.exist?(filename)
      raise "Файл не найден"
    end
    CSV.foreach(filename, headers: true, col_sep: ";") do |row|
      begin
        add_movie(
          title: row["Название"],
          genre: row["Жанр"],
          director: row["Режиссёр"],
          year: row["Год"].to_i,
          watched: row["Просмотрен"] == "Да",
          rating: row["Оценка"].to_i,
          notes: row["Заметки"]
        )
      rescue => e
        puts "Ошибка импорта строки: #{e}"
      end
    end
  end
end

def print_movie(movie)
  status = movie.watched ? "✅ Просмотрен" : "⏳ Не просмотрен"
  puts "##{movie.id} - #{movie.title} (#{movie.year})"
  puts "   Жанр: #{movie.genre}, Режиссёр: #{movie.director}"
  puts "   #{status}, Оценка: #{movie.rating}/10"
  puts "   Заметки: #{movie.notes}" unless movie.notes.empty?
  puts "   Добавлен: #{movie.added_date}"
end

def main
  randomizer = MovieRandomizer.new
  randomizer.load_from_file

  loop do
    puts "\n===== ГЕНЕРАТОР СЛУЧАЙНОГО ФИЛЬМА (Ruby) ====="
    puts "1. Добавить фильм"
    puts "2. Показать все фильмы"
    puts "3. Рекомендовать случайный фильм"
    puts "4. Рекомендовать по жанру"
    puts "5. Показать непросмотренные фильмы"
    puts "6. Отметить фильм как просмотренный"
    puts "7. Редактировать фильм"
    puts "8. Удалить фильм"
    puts "9. Показать статистику"
    puts "10. Сохранить в файл"
    puts "11. Загрузить из файла"
    puts "12. Экспорт в CSV"
    puts "13. Импорт из CSV"
    puts "0. Выход"
    print "Выберите действие: "
    choice = gets.chomp

    case choice
    when "0"
      break
    when "1"
      print "Название: "
      title = gets.chomp
      next if title.empty?
      print "Жанр: "
      genre = gets.chomp
      next if genre.empty?
      print "Режиссёр (необязательно): "
      director = gets.chomp
      print "Год выпуска: "
      year = gets.chomp.to_i
      print "Статус (1-просмотрен, 0-нет): "
      watched = gets.chomp == "1"
      print "Оценка (1-10): "
      rating = gets.chomp.to_i
      print "Заметки (необязательно): "
      notes = gets.chomp
      begin
        movie = randomizer.add_movie(title, genre, director, year, watched, rating, notes)
        puts "Фильм добавлен с ID #{movie.id}"
      rescue => e
        puts "Ошибка: #{e.message}"
      end
    when "2"
      if randomizer.movies.empty?
        puts "Нет фильмов."
      else
        randomizer.movies.each { |m| print_movie(m) }
      end
    when "3"
      movie = randomizer.random_movie
      if movie.nil?
        puts "Нет фильмов в коллекции."
      else
        puts "\n🎬 Рекомендую посмотреть:"
        print_movie(movie)
      end
    when "4"
      print "Введите жанр: "
      genre = gets.chomp
      if genre.empty?
        puts "Введите жанр."
        next
      end
      movie = randomizer.random_movie(genre)
      if movie.nil?
        puts "Нет фильмов в жанре '#{genre}'."
      else
        puts "\n🎬 Рекомендую посмотреть в жанре #{genre}:"
        print_movie(movie)
      end
    when "5"
      unwatched = randomizer.unwatched
      if unwatched.empty?
        puts "Нет непросмотренных фильмов."
      else
        unwatched.each { |m| print_movie(m) }
      end
    when "6"
      print "Введите ID фильма: "
      id = gets.chomp.to_i
      movie = randomizer.find_movie(id)
      unless movie
        puts "Фильм не найден."
        next
      end
      if movie.watched
        puts "Фильм уже отмечен как просмотренный."
      else
        movie.watched = true
        puts "Фильм отмечен как просмотренный."
      end
    when "7"
      print "Введите ID фильма для редактирования: "
      id = gets.chomp.to_i
      movie = randomizer.find_movie(id)
      unless movie
        puts "Фильм не найден."
        next
      end
      puts "Оставьте поле пустым, чтобы не менять."
      print "Название (#{movie.title}): "
      new_title = gets.chomp
      print "Жанр (#{movie.genre}): "
      new_genre = gets.chomp
      print "Режиссёр (#{movie.director}): "
      new_director = gets.chomp
      print "Год (#{movie.year}): "
      new_year = gets.chomp
      print "Статус (1-просмотрен, 0-нет) сейчас: #{movie.watched ? '1' : '0'}: "
      new_watched = gets.chomp
      print "Оценка (#{movie.rating}): "
      new_rating = gets.chomp
      print "Заметки (#{movie.notes}): "
      new_notes = gets.chomp
      updates = {}
      updates[:title] = new_title unless new_title.empty?
      updates[:genre] = new_genre unless new_genre.empty?
      updates[:director] = new_director unless new_director.empty?
      unless new_year.empty?
        updates[:year] = new_year.to_i
      end
      unless new_watched.empty?
        updates[:watched] = new_watched == "1"
      end
      unless new_rating.empty?
        updates[:rating] = new_rating.to_i
      end
      updates[:notes] = new_notes unless new_notes.empty?
      if randomizer.edit_movie(id, **updates)
        puts "Фильм обновлён."
      else
        puts "Ошибка обновления."
      end
    when "8"
      print "Введите ID фильма для удаления: "
      id = gets.chomp.to_i
      if randomizer.delete_movie(id)
        puts "Фильм удалён."
      else
        puts "Фильм не найден."
      end
    when "9"
      stats = randomizer.stats
      puts "\n=== СТАТИСТИКА ==="
      puts "Всего фильмов: #{stats[:total]}"
      puts "Просмотрено: #{stats[:watched]}"
      puts "Не просмотрено: #{stats[:unwatched]}"
      puts "Средняя оценка (просмотренные): #{'%.2f' % stats[:avg_rating]}"
      puts "По жанрам:"
      stats[:genres].each { |g, c| puts "  #{g}: #{c}" }
    when "10"
      randomizer.save_to_file
      puts "Сохранено."
    when "11"
      randomizer.load_from_file
      puts "Загружено."
    when "12"
      randomizer.export_csv
      puts "Экспортировано в movies_export.csv"
    when "13"
      begin
        randomizer.import_csv
        puts "Импортировано из movies_export.csv"
      rescue => e
        puts "Ошибка импорта: #{e}"
      end
    else
      puts "Неизвестная команда."
    end
  end
end

main if __FILE__ == $0
