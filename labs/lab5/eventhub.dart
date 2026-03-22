import 'package:flutter/material.dart';

void main() {
  runApp(const EventHubApp());
}

class EventHubApp extends StatelessWidget {
  const EventHubApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'EventHub',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.deepPurple,
        ),
        useMaterial3: true,
      ),
      home: const EventListScreen(),
    );
  }
}

// ==============================
// МОДЕЛЬ ДАННЫХ
// ==============================

class EventCategory {
  final String name;
  final IconData icon;
  final Color color;

  const EventCategory({
    required this.name,
    required this.icon,
    required this.color,
  });
}

class Event {
  final String id; // уникальный идентификатор
  final String title;
  final String description;
  final String location;
  final EventCategory category;
  final DateTime date;
  final TimeOfDay time;
  final List<String> participants;
  final String emoji;

  Event({
    required this.id,
    required this.title,
    required this.description,
    required this.location,
    required this.category,
    required this.date,
    required this.time,
    required this.participants,
    required this.emoji,
  });
}

// ==============================
// КАТЕГОРИИ
// ==============================

const categories = [
  EventCategory(
    name: 'Учёба',
    icon: Icons.school,
    color: Colors.blue,
  ),
  EventCategory(
    name: 'Спорт',
    icon: Icons.sports_soccer,
    color: Colors.green,
  ),
  EventCategory(
    name: 'Развлечения',
    icon: Icons.celebration,
    color: Colors.orange,
  ),
  EventCategory(
    name: 'Работа',
    icon: Icons.work,
    color: Colors.red,
  ),
  EventCategory(
    name: 'Личное',
    icon: Icons.favorite,
    color: Colors.pink,
  ),
];

// ==============================
// НАЧАЛЬНЫЕ ДАННЫЕ
// ==============================

List<Event> events = [
  Event(
    id: '1',
    title: 'Лекция по Flutter',
    description:
        'Лабораторная работа №4. Создание приложения EventHub с использованием GridView, BottomSheet и других виджетов.',
    location: 'Аудитория 305',
    category: categories[0],
    date: DateTime.now(),
    time: const TimeOfDay(hour: 9, minute: 0),
    participants: [
      'Иванов А.',
      'Петрова Б.',
      'Сидоров В.',
    ],
    emoji: '📚',
  ),
  Event(
    id: '2',
    title: 'Футбол с друзьями',
    description: 'Товарищеский матч 5 на 5. Не забудь форму и воду!',
    location: 'Стадион «Спартак»',
    category: categories[1],
    date: DateTime.now().add(const Duration(days: 1)),
    time: const TimeOfDay(hour: 18, minute: 30),
    participants: [
      'Команда А',
      'Команда Б',
    ],
    emoji: '⚽',
  ),
  Event(
    id: '3',
    title: 'Кинопремьера',
    description: 'Новый фильм в IMAX. Билеты уже куплены, ряд 7.',
    location: 'Кинотеатр «Синема Парк»',
    category: categories[2],
    date: DateTime.now().add(const Duration(days: 2)),
    time: const TimeOfDay(hour: 20, minute: 0),
    participants: [
      'Аня',
      'Максим',
      'Даша',
    ],
    emoji: '🎬',
  ),
  Event(
    id: '4',
    title: 'Митап по мобильной разработке',
    description:
        'Доклады: Compose vs Flutter, архитектура чистого кода, CI/CD для мобильных приложений.',
    location: 'Коворкинг «Точка кипения»',
    category: categories[3],
    date: DateTime.now().add(const Duration(days: 3)),
    time: const TimeOfDay(hour: 19, minute: 0),
    participants: [
      'Спикер 1',
      'Спикер 2',
      '~50 участников',
    ],
    emoji: '💻',
  ),
  Event(
    id: '5',
    title: 'День рождения Маши',
    description: 'Собираемся у Маши дома. Подарок: книга по Dart.',
    location: 'ул. Ленина, 42',
    category: categories[4],
    date: DateTime.now().add(const Duration(days: 5)),
    time: const TimeOfDay(hour: 17, minute: 0),
    participants: [
      'Маша',
      'Ваня',
      'Катя',
      'Олег',
      'Лиза',
    ],
    emoji: '🎂',
  ),
  Event(
    id: '6',
    title: 'Защита курсовой',
    description:
        'Финальная защита курсовой работы по дисциплине «Мобильная разработка».',
    location: 'Аудитория 112',
    category: categories[0],
    date: DateTime.now().add(const Duration(days: 7)),
    time: const TimeOfDay(hour: 10, minute: 0),
    participants: [
      'Группа ИСТ-21',
      'Преподаватель',
    ],
    emoji: '🎓',
  ),
];

// ==============================
// ГЛАВНЫЙ ЭКРАН (с поиском и сортировкой)
// ==============================

class EventListScreen extends StatefulWidget {
  const EventListScreen({super.key});

  @override
  State<EventListScreen> createState() => _EventListScreenState();
}

class _EventListScreenState extends State<EventListScreen> {
  String _selectedCategory = 'Все';
  String _searchQuery = '';
  bool _isSearching = false;
  final TextEditingController _searchController = TextEditingController();
  String _sortCriterion = 'date'; // 'date', 'title', 'category'

  List<Event> get _filteredEvents {
    // Сначала фильтруем по категории
    Iterable<Event> filtered = events;
    if (_selectedCategory != 'Все') {
      filtered = filtered.where((e) => e.category.name == _selectedCategory);
    }
    // Затем по поисковому запросу
    if (_searchQuery.isNotEmpty) {
      filtered = filtered.where(
          (e) => e.title.toLowerCase().contains(_searchQuery.toLowerCase()));
    }
    // Сортируем
    List<Event> sorted = filtered.toList();
    switch (_sortCriterion) {
      case 'title':
        sorted.sort((a, b) => a.title.compareTo(b.title));
        break;
      case 'category':
        sorted.sort((a, b) => a.category.name.compareTo(b.category.name));
        break;
      case 'date':
      default:
        sorted.sort((a, b) => a.date.compareTo(b.date));
        break;
    }
    return sorted;
  }

  String _formatDate(DateTime d) {
    const months = [
      '',
      'янв',
      'фев',
      'мар',
      'апр',
      'май',
      'июн',
      'июл',
      'авг',
      'сен',
      'окт',
      'ноя',
      'дек'
    ];
    return '${d.day} ${months[d.month]}';
  }

  String _formatTime(TimeOfDay t) {
    final h = t.hour.toString().padLeft(2, '0');
    final m = t.minute.toString().padLeft(2, '0');
    return '$h:$m';
  }

  void _startSearch() {
    setState(() {
      _isSearching = true;
    });
  }

  void _stopSearch() {
    setState(() {
      _isSearching = false;
      _searchQuery = '';
      _searchController.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: _isSearching
            ? TextField(
                controller: _searchController,
                autofocus: true,
                decoration: const InputDecoration(
                  hintText: 'Поиск событий...',
                  border: InputBorder.none,
                ),
                onChanged: (value) {
                  setState(() {
                    _searchQuery = value;
                  });
                },
              )
            : const Text('EventHub'),
        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
        actions: _isSearching
            ? [
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: _stopSearch,
                ),
              ]
            : [
                IconButton(
                  icon: const Icon(Icons.search),
                  onPressed: _startSearch,
                ),
                PopupMenuButton<String>(
                  icon: const Icon(Icons.sort),
                  onSelected: (value) {
                    setState(() {
                      _sortCriterion = value;
                    });
                  },
                  itemBuilder: (context) => [
                    const PopupMenuItem(
                      value: 'date',
                      child: Text('По дате'),
                    ),
                    const PopupMenuItem(
                      value: 'title',
                      child: Text('По названию'),
                    ),
                    const PopupMenuItem(
                      value: 'category',
                      child: Text('По категории'),
                    ),
                  ],
                ),
                IconButton(
                  icon: const Icon(Icons.pie_chart),
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const StatisticsScreen(),
                      ),
                    ).then((_) => setState(() {}));
                  },
                ),
              ],
      ),
      body: Column(
        children: [
          // Фильтр по категориям
          Container(
            padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
            color: Theme.of(context).colorScheme.surface,
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: Row(
                children: [
                  // Чип «Все»
                  Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: ChoiceChip(
                      label: const Text('Все'),
                      avatar: const Icon(Icons.apps, size: 18),
                      selected: _selectedCategory == 'Все',
                      onSelected: (selected) {
                        setState(() {
                          _selectedCategory = 'Все';
                        });
                      },
                    ),
                  ),
                  // Чипы категорий
                  ...categories.map(
                    (cat) => Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: ChoiceChip(
                        label: Text(cat.name),
                        avatar: Icon(cat.icon, size: 18),
                        selected: _selectedCategory == cat.name,
                        onSelected: (selected) {
                          setState(() {
                            _selectedCategory = selected ? cat.name : 'Все';
                          });
                        },
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          // Строка статистики
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Row(
              children: [
                Icon(Icons.event_note, size: 18, color: Colors.grey[600]),
                const SizedBox(width: 6),
                Text(
                  _selectedCategory == 'Все'
                      ? 'Всего событий: ${events.length}'
                      : '$_selectedCategory: ${_filteredEvents.length} из ${events.length}',
                  style: TextStyle(
                    color: Colors.grey[600],
                    fontSize: 14,
                  ),
                ),
              ],
            ),
          ),
          // Сетка событий
          Expanded(
            child: _filteredEvents.isEmpty
                ? Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.event_busy,
                            size: 64, color: Colors.grey[300]),
                        const SizedBox(height: 12),
                        Text(
                          'Нет событий',
                          style: TextStyle(
                            color: Colors.grey[400],
                            fontSize: 18,
                          ),
                        ),
                      ],
                    ),
                  )
                : GridView.count(
                    crossAxisCount: 2,
                    padding: const EdgeInsets.all(12),
                    crossAxisSpacing: 12,
                    mainAxisSpacing: 12,
                    childAspectRatio: 0.85,
                    children: _filteredEvents
                        .map((event) => _buildEventCard(event))
                        .toList(),
                  ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showAddEventSheet(context),
        icon: const Icon(Icons.add),
        label: const Text('Событие'),
      ),
    );
  }

  Widget _buildEventCard(Event event) {
    return Dismissible(
      key: ValueKey(event.id),
      direction: DismissDirection.endToStart,
      background: Container(
        alignment: Alignment.centerRight,
        padding: const EdgeInsets.only(right: 20),
        decoration: BoxDecoration(
          color: Colors.red[400],
          borderRadius: BorderRadius.circular(16),
        ),
        child: const Icon(Icons.delete, color: Colors.white, size: 32),
      ),
      onDismissed: (direction) {
        setState(() {
          events.remove(event);
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('${event.title} удалено'),
            action: SnackBarAction(
              label: 'Отменить',
              onPressed: () {
                setState(() {
                  events.add(event);
                });
              },
            ),
          ),
        );
      },
      child: GestureDetector(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => EventDetailScreen(event: event),
            ),
          ).then((_) => setState(() {})); // обновление после возврата
        },
        child: Card(
          clipBehavior: Clip.antiAlias,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Цветной заголовок
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      event.category.color.withOpacity(0.7),
                      event.category.color.withOpacity(0.4),
                    ],
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(event.emoji, style: const TextStyle(fontSize: 32)),
                    const SizedBox(height: 4),
                    Text(
                      event.title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.bold,
                        fontSize: 15,
                      ),
                    ),
                  ],
                ),
              ),
              // Информация
              Padding(
                padding: const EdgeInsets.all(10),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.calendar_today,
                            size: 14, color: Colors.grey[600]),
                        const SizedBox(width: 4),
                        Text(
                          _formatDate(event.date),
                          style:
                              TextStyle(fontSize: 12, color: Colors.grey[600]),
                        ),
                        const SizedBox(width: 8),
                        Icon(Icons.access_time,
                            size: 14, color: Colors.grey[600]),
                        const SizedBox(width: 4),
                        Text(
                          _formatTime(event.time),
                          style:
                              TextStyle(fontSize: 12, color: Colors.grey[600]),
                        ),
                      ],
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        Icon(Icons.location_on,
                            size: 14, color: Colors.grey[600]),
                        const SizedBox(width: 4),
                        Expanded(
                          child: Text(
                            event.location,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                                fontSize: 12, color: Colors.grey[600]),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showAddEventSheet(BuildContext context) {
    final titleCtrl = TextEditingController();
    final locationCtrl = TextEditingController();
    EventCategory selectedCat = categories[0];
    DateTime selectedDate = DateTime.now();
    TimeOfDay selectedTime = TimeOfDay.now();

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setSheetState) {
            return Padding(
              padding: EdgeInsets.only(
                left: 20,
                right: 20,
                top: 20,
                bottom: MediaQuery.of(ctx).viewInsets.bottom + 20,
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.add_circle, color: Colors.deepPurple),
                      const SizedBox(width: 8),
                      const Text(
                        'Новое событие',
                        style: TextStyle(
                            fontSize: 20, fontWeight: FontWeight.bold),
                      ),
                      const Spacer(),
                      IconButton(
                        onPressed: () => Navigator.pop(ctx),
                        icon: const Icon(Icons.close),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: titleCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Название',
                      prefixIcon: Icon(Icons.title),
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: locationCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Место',
                      prefixIcon: Icon(Icons.location_on),
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 12),
                  DropdownButtonFormField<EventCategory>(
                    value: selectedCat,
                    decoration: const InputDecoration(
                      labelText: 'Категория',
                      prefixIcon: Icon(Icons.category),
                      border: OutlineInputBorder(),
                    ),
                    items: categories
                        .map((cat) => DropdownMenuItem(
                              value: cat,
                              child: Row(
                                children: [
                                  Icon(cat.icon, size: 20, color: cat.color),
                                  const SizedBox(width: 8),
                                  Text(cat.name),
                                ],
                              ),
                            ))
                        .toList(),
                    onChanged: (val) {
                      if (val != null) {
                        setSheetState(() {
                          selectedCat = val;
                        });
                      }
                    },
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: () async {
                            final d = await showDatePicker(
                              context: ctx,
                              initialDate: selectedDate,
                              firstDate: DateTime.now(),
                              lastDate:
                                  DateTime.now().add(const Duration(days: 365)),
                            );
                            if (d != null) {
                              setSheetState(() {
                                selectedDate = d;
                              });
                            }
                          },
                          icon: const Icon(Icons.calendar_today),
                          label: Text(
                              '${selectedDate.day}.${selectedDate.month}.${selectedDate.year}'),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: () async {
                            final t = await showTimePicker(
                              context: ctx,
                              initialTime: selectedTime,
                            );
                            if (t != null) {
                              setSheetState(() {
                                selectedTime = t;
                              });
                            }
                          },
                          icon: const Icon(Icons.access_time),
                          label: Text(
                              '${selectedTime.hour.toString().padLeft(2, "0")}:${selectedTime.minute.toString().padLeft(2, "0")}'),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  FilledButton.icon(
                    onPressed: () {
                      if (titleCtrl.text.isNotEmpty) {
                        setState(() {
                          events.add(Event(
                            id: DateTime.now()
                                .millisecondsSinceEpoch
                                .toString(),
                            title: titleCtrl.text,
                            description: 'Без описания',
                            location: locationCtrl.text.isNotEmpty
                                ? locationCtrl.text
                                : 'Не указано',
                            category: selectedCat,
                            date: selectedDate,
                            time: selectedTime,
                            participants: [],
                            emoji: '📌',
                          ));
                        });
                        Navigator.pop(ctx);
                      }
                    },
                    icon: const Icon(Icons.check),
                    label: const Text('Создать событие'),
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }
}

// ==============================
// ЭКРАН ДЕТАЛЕЙ СОБЫТИЯ (с редактированием)
// ==============================

class EventDetailScreen extends StatefulWidget {
  final Event event;

  const EventDetailScreen({super.key, required this.event});

  @override
  State<EventDetailScreen> createState() => _EventDetailScreenState();
}

class _EventDetailScreenState extends State<EventDetailScreen> {
  late Event _event;

  @override
  void initState() {
    super.initState();
    _event = widget.event;
  }

  void _editEvent(BuildContext context) {
    final titleCtrl = TextEditingController(text: _event.title);
    final locationCtrl = TextEditingController(text: _event.location);
    EventCategory selectedCat = _event.category;
    DateTime selectedDate = _event.date;
    TimeOfDay selectedTime = _event.time;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setSheetState) {
            return Padding(
              padding: EdgeInsets.only(
                left: 20,
                right: 20,
                top: 20,
                bottom: MediaQuery.of(ctx).viewInsets.bottom + 20,
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.edit, color: Colors.deepPurple),
                      const SizedBox(width: 8),
                      const Text(
                        'Редактировать событие',
                        style: TextStyle(
                            fontSize: 20, fontWeight: FontWeight.bold),
                      ),
                      const Spacer(),
                      IconButton(
                        onPressed: () => Navigator.pop(ctx),
                        icon: const Icon(Icons.close),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: titleCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Название',
                      prefixIcon: Icon(Icons.title),
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: locationCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Место',
                      prefixIcon: Icon(Icons.location_on),
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 12),
                  DropdownButtonFormField<EventCategory>(
                    value: selectedCat,
                    decoration: const InputDecoration(
                      labelText: 'Категория',
                      prefixIcon: Icon(Icons.category),
                      border: OutlineInputBorder(),
                    ),
                    items: categories
                        .map((cat) => DropdownMenuItem(
                              value: cat,
                              child: Row(
                                children: [
                                  Icon(cat.icon, size: 20, color: cat.color),
                                  const SizedBox(width: 8),
                                  Text(cat.name),
                                ],
                              ),
                            ))
                        .toList(),
                    onChanged: (val) {
                      if (val != null) {
                        setSheetState(() {
                          selectedCat = val;
                        });
                      }
                    },
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: () async {
                            final d = await showDatePicker(
                              context: ctx,
                              initialDate: selectedDate,
                              firstDate: DateTime.now()
                                  .subtract(const Duration(days: 365)),
                              lastDate:
                                  DateTime.now().add(const Duration(days: 365)),
                            );
                            if (d != null) {
                              setSheetState(() {
                                selectedDate = d;
                              });
                            }
                          },
                          icon: const Icon(Icons.calendar_today),
                          label: Text(
                              '${selectedDate.day}.${selectedDate.month}.${selectedDate.year}'),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: () async {
                            final t = await showTimePicker(
                              context: ctx,
                              initialTime: selectedTime,
                            );
                            if (t != null) {
                              setSheetState(() {
                                selectedTime = t;
                              });
                            }
                          },
                          icon: const Icon(Icons.access_time),
                          label: Text(
                              '${selectedTime.hour.toString().padLeft(2, "0")}:${selectedTime.minute.toString().padLeft(2, "0")}'),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  FilledButton.icon(
                    onPressed: () {
                      if (titleCtrl.text.isNotEmpty) {
                        // Найти индекс события и заменить
                        final index =
                            events.indexWhere((e) => e.id == _event.id);
                        if (index != -1) {
                          final updatedEvent = Event(
                            id: _event.id,
                            title: titleCtrl.text,
                            description: _event.description,
                            location: locationCtrl.text.isNotEmpty
                                ? locationCtrl.text
                                : 'Не указано',
                            category: selectedCat,
                            date: selectedDate,
                            time: selectedTime,
                            participants: _event.participants,
                            emoji: _event.emoji,
                          );
                          setState(() {
                            events[index] = updatedEvent;
                            _event = updatedEvent; // обновить локально
                          });
                        }
                        Navigator.pop(ctx);
                      }
                    },
                    icon: const Icon(Icons.save),
                    label: const Text('Сохранить'),
                  ),
                ],
              ),
            );
          },
        );
      },
    ).then((_) => setState(() {})); // обновить экран деталей после закрытия
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_event.title),
        backgroundColor: _event.category.color.withOpacity(0.3),
        actions: [
          IconButton(
            icon: const Icon(Icons.edit),
            onPressed: () => _editEvent(context),
          ),
        ],
      ),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Баннер
            Container(
              height: 180,
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    _event.category.color.withOpacity(0.6),
                    _event.category.color.withOpacity(0.2),
                  ],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
              ),
              child: Stack(
                children: [
                  Positioned(
                    right: 20,
                    bottom: 10,
                    child: Text(
                      _event.emoji,
                      style: TextStyle(
                        fontSize: 100,
                        color: Colors.white.withOpacity(0.3),
                      ),
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 10, vertical: 4),
                          decoration: BoxDecoration(
                            color: Colors.white24,
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(_event.category.icon,
                                  size: 16, color: Colors.white),
                              const SizedBox(width: 4),
                              Text(
                                _event.category.name,
                                style: const TextStyle(
                                    color: Colors.white, fontSize: 13),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _event.title,
                          style: const TextStyle(
                            fontSize: 26,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            // Дата, время, место
            Padding(
              padding: const EdgeInsets.all(16),
              child: Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      _infoRow(
                        Icons.calendar_today,
                        'Дата',
                        '${_event.date.day}.${_event.date.month}.${_event.date.year}',
                        _event.category.color,
                      ),
                      const Divider(),
                      _infoRow(
                        Icons.access_time,
                        'Время',
                        '${_event.time.hour.toString().padLeft(2, "0")}:${_event.time.minute.toString().padLeft(2, "0")}',
                        _event.category.color,
                      ),
                      const Divider(),
                      _infoRow(
                        Icons.location_on,
                        'Место',
                        _event.location,
                        _event.category.color,
                      ),
                    ],
                  ),
                ),
              ),
            ),
            // Раскрываемое описание
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Card(
                clipBehavior: Clip.antiAlias,
                child: ExpansionTile(
                  leading:
                      Icon(Icons.description, color: _event.category.color),
                  title: const Text(
                    'Описание',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  initiallyExpanded: true,
                  children: [
                    Padding(
                      padding: const EdgeInsets.all(16),
                      child: Text(
                        _event.description,
                        style: const TextStyle(fontSize: 15, height: 1.5),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 8),
            // Раскрываемые участники
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Card(
                clipBehavior: Clip.antiAlias,
                child: ExpansionTile(
                  leading: Icon(Icons.people, color: _event.category.color),
                  title: Text(
                    'Участники (${_event.participants.length})',
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                  children: [
                    ..._event.participants.map(
                      (name) => ListTile(
                        leading: CircleAvatar(
                          backgroundColor:
                              _event.category.color.withOpacity(0.2),
                          child: Text(
                            name[0],
                            style: TextStyle(
                              color: _event.category.color,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                        title: Text(name),
                      ),
                    ),
                    const SizedBox(height: 8),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _infoRow(IconData icon, String label, String value, Color color) {
    return Row(
      children: [
        Icon(icon, color: color, size: 22),
        const SizedBox(width: 12),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              label,
              style: TextStyle(fontSize: 12, color: Colors.grey[500]),
            ),
            Text(
              value,
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
            ),
          ],
        ),
      ],
    );
  }
}

// ==============================
// ЭКРАН СТАТИСТИКИ (исправлен)
// ==============================

class StatisticsScreen extends StatelessWidget {
  const StatisticsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final totalEvents = events.length;

    // Подсчет событий по категориям
    Map<EventCategory, int> categoryCounts = {};
    for (var cat in categories) {
      categoryCounts[cat] = events.where((e) => e.category == cat).length;
    }

    // Ближайшее событие
    final now = DateTime.now();
    final upcomingEvents = List<Event>.from(events)
      ..sort((a, b) => a.date.compareTo(b.date));

    Event? nextEvent;
    if (upcomingEvents.isNotEmpty) {
      nextEvent = upcomingEvents.firstWhere(
        (e) =>
            e.date.isAfter(now) ||
            (e.date.isAtSameMomentAs(now) &&
                e.time.hour * 60 + e.time.minute >= now.hour * 60 + now.minute),
        orElse: () => upcomingEvents.first,
      );
    }

    // Ближайшие 3 события
    final nextThree = upcomingEvents.take(3).toList();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Статистика'),
        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Общее количество
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    const Icon(Icons.event, size: 40, color: Colors.deepPurple),
                    const SizedBox(width: 16),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Всего событий',
                          style: TextStyle(fontSize: 14, color: Colors.grey),
                        ),
                        Text(
                          '$totalEvents',
                          style: const TextStyle(
                              fontSize: 32, fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            // Ближайшее событие
            if (nextEvent != null)
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Ближайшее событие',
                        style: TextStyle(fontSize: 14, color: Colors.grey),
                      ),
                      const SizedBox(height: 8),
                      ListTile(
                        leading: CircleAvatar(
                          backgroundColor: nextEvent.category.color,
                          child: Text(nextEvent.emoji),
                        ),
                        title: Text(nextEvent.title),
                        subtitle: Text(
                            '${nextEvent.date.day}.${nextEvent.date.month}.${nextEvent.date.year}'),
                      ),
                    ],
                  ),
                ),
              ),
            const SizedBox(height: 16),
            // Статистика по категориям
            const Text(
              'Категории',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            ...categories.map((cat) {
              int count = categoryCounts[cat] ?? 0;
              double fraction = totalEvents > 0 ? count / totalEvents : 0;
              return Card(
                margin: const EdgeInsets.symmetric(vertical: 4),
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Row(
                    children: [
                      // Круговой прогресс
                      SizedBox(
                        width: 60,
                        height: 60,
                        child: Stack(
                          alignment: Alignment.center,
                          children: [
                            SizedBox(
                              width: 50,
                              height: 50,
                              child: CircularProgressIndicator(
                                value: fraction,
                                strokeWidth: 6,
                                color: cat.color,
                                backgroundColor: Colors.grey[200],
                              ),
                            ),
                            Text(
                              '$count',
                              style: const TextStyle(
                                  fontWeight: FontWeight.bold, fontSize: 16),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              cat.name,
                              style: const TextStyle(
                                  fontSize: 16, fontWeight: FontWeight.w500),
                            ),
                            const SizedBox(height: 4),
                            LinearProgressIndicator(
                              value: fraction,
                              color: cat.color,
                              backgroundColor: Colors.grey[200],
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }),
            const SizedBox(height: 16),
            // Ближайшие 3 события
            const Text(
              'Ближайшие события',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            ...nextThree.map((event) {
              return Card(
                margin: const EdgeInsets.symmetric(vertical: 4),
                child: ExpansionTile(
                  leading: CircleAvatar(
                    backgroundColor: event.category.color.withOpacity(0.2),
                    child: Text(event.emoji),
                  ),
                  title: Text(event.title),
                  subtitle: Text(
                      '${event.date.day}.${event.date.month}.${event.date.year}'),
                  children: [
                    Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Описание: ${event.description}'),
                          const SizedBox(height: 4),
                          Text('Место: ${event.location}'),
                          const SizedBox(height: 4),
                          Text('Участников: ${event.participants.length}'),
                        ],
                      ),
                    ),
                  ],
                ),
              );
            }),
          ],
        ),
      ),
    );
  }
}
