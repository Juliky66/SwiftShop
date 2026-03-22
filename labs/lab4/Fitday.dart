import 'package:flutter/material.dart';
import 'dart:async';

void main() {
  runApp(const FitDayApp());
}

// ==============================
// Главный класс приложения
// ==============================
class FitDayApp extends StatelessWidget {
  const FitDayApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'FitDay',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.teal),
        useMaterial3: true,
      ),
      home: const HomeScreen(),
    );
  }
}

// ==============================
// Модель данных
// ==============================
class Exercise {
  String name;
  String emoji;
  int sets;
  int reps;
  int calories;
  bool isDone;

  Exercise({
    required this.name,
    required this.emoji,
    required this.sets,
    required this.reps,
    required this.calories,
    this.isDone = false,
  });
}

// ==============================
// Данные тренировки
// ==============================
List<Exercise> todayWorkout = [
  Exercise(name: 'Отжимания', emoji: '💪', sets: 3, reps: 15, calories: 50),
  Exercise(name: 'Приседания', emoji: '🧎', sets: 4, reps: 20, calories: 70),
  Exercise(name: 'Планка', emoji: '🧘', sets: 3, reps: 1, calories: 40),
  Exercise(name: 'Бёрпи', emoji: '🔥', sets: 3, reps: 10, calories: 90),
  Exercise(name: 'Скручивания', emoji: '🏋', sets: 3, reps: 20, calories: 45),
  Exercise(name: 'Выпады', emoji: '🏃', sets: 3, reps: 12, calories: 60),
];

// Цель по калориям
int calorieGoal = 300;

// ==============================
// История тренировок
// ==============================
List<Map<String, dynamic>> history = [];

// ==============================
// Главный экран с 4 вкладками
// ==============================
class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 5, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  void _refresh() {
    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('FitDay'),
        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          labelColor: Colors.black,
          unselectedLabelColor: Colors.grey,
          indicatorColor: Colors.teal,
          tabs: const [
            Tab(icon: Icon(Icons.fitness_center), text: 'Тренировка'),
            Tab(icon: Icon(Icons.bar_chart), text: 'Прогресс'),
            Tab(icon: Icon(Icons.settings), text: 'Настройки'),
            Tab(icon: Icon(Icons.history), text: 'История'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          WorkoutTab(onChanged: _refresh),
          const ProgressTab(),
          SettingsTab(onChanged: _refresh),
          const HistoryTab(),
        ],
      ),
    );
  }
}

// ==============================
// Вкладка 1: Тренировка
// ==============================
class WorkoutTab extends StatefulWidget {
  final VoidCallback onChanged;
  const WorkoutTab({super.key, required this.onChanged});

  @override
  State<WorkoutTab> createState() => _WorkoutTabState();
}

class _WorkoutTabState extends State<WorkoutTab> {
  void _addExercise() {
    final nameController = TextEditingController();
    final setsController = TextEditingController(text: '3');
    final repsController = TextEditingController(text: '10');

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Новое упражнение'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: nameController,
              decoration: const InputDecoration(
                labelText: 'Название',
                hintText: 'Например: Подтягивания',
              ),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: setsController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Подходы'),
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: TextField(
                    controller: repsController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Повторения'),
                  ),
                ),
              ],
            ),
          ],
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Отмена')),
          FilledButton(
            onPressed: () {
              if (nameController.text.isNotEmpty) {
                setState(() {
                  todayWorkout.add(Exercise(
                    name: nameController.text,
                    emoji: '⭐',
                    sets: int.tryParse(setsController.text) ?? 3,
                    reps: int.tryParse(repsController.text) ?? 10,
                    calories: 30,
                  ));
                });
                widget.onChanged();
                Navigator.pop(context);
              }
            },
            child: const Text('Добавить'),
          ),
        ],
      ),
    );
  }

  void _deleteExercise(int index) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Удалить упражнение?'),
        content: const Text('Вы уверены, что хотите удалить это упражнение?'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Отмена')),
          FilledButton(
            onPressed: () {
              setState(() {
                todayWorkout.removeAt(index);
              });
              widget.onChanged();
              Navigator.pop(context);
            },
            child: const Text('Удалить'),
          ),
        ],
      ),
    );
  }

  void _finishDay() {
    final doneCount = todayWorkout.where((e) => e.isDone).length;
    final totalCalories = todayWorkout
        .where((e) => e.isDone)
        .fold(0, (sum, e) => sum + e.calories);

    history.add({
      'date': DateTime.now().toString().substring(0, 10),
      'exercises': doneCount,
      'calories': totalCalories,
    });

    for (var e in todayWorkout) e.isDone = false;
    setState(() {});
    widget.onChanged();

    ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('День сохранён в историю!')));
  }

  @override
  Widget build(BuildContext context) {
    final doneCount = todayWorkout.where((e) => e.isDone).length;

    return Scaffold(
      body: Column(
        children: [
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(20),
            color:
                Theme.of(context).colorScheme.primaryContainer.withOpacity(0.3),
            child: Column(
              children: [
                Text('Сегодня: $doneCount из ${todayWorkout.length}',
                    style: const TextStyle(
                        fontSize: 20, fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: LinearProgressIndicator(
                    value: todayWorkout.isEmpty
                        ? 0
                        : doneCount / todayWorkout.length,
                    minHeight: 12,
                    backgroundColor: Colors.grey[300],
                  ),
                ),
                const SizedBox(height: 8),
                FilledButton(
                    onPressed: _finishDay, child: const Text('Завершить день')),
              ],
            ),
          ),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(12),
              itemCount: todayWorkout.length,
              itemBuilder: (context, index) {
                final ex = todayWorkout[index];
                return Card(
                  color: ex.isDone ? Colors.teal.withOpacity(0.1) : null,
                  margin: const EdgeInsets.only(bottom: 8),
                  child: ListTile(
                    leading:
                        Text(ex.emoji, style: const TextStyle(fontSize: 32)),
                    title: Text(
                      ex.name,
                      style: TextStyle(
                          fontWeight: FontWeight.bold,
                          decoration:
                              ex.isDone ? TextDecoration.lineThrough : null),
                    ),
                    subtitle: Text(
                        '${ex.sets} подходов × ${ex.reps} повт. | ${ex.calories} ккал'),
                    trailing: Switch(
                      value: ex.isDone,
                      onChanged: (val) {
                        setState(() {
                          ex.isDone = val;
                        });
                        widget.onChanged();
                      },
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                            builder: (_) => TimerScreen(exercise: ex)),
                      );
                    },
                    onLongPress: () => _deleteExercise(index),
                  ),
                );
              },
            ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _addExercise,
        child: const Icon(Icons.add),
      ),
    );
  }
}

// ==============================
// Вкладка 2: Прогресс
// ==============================
class ProgressTab extends StatelessWidget {
  const ProgressTab({super.key});

  @override
  Widget build(BuildContext context) {
    final doneCount = todayWorkout.where((e) => e.isDone).length;
    final totalCalories = todayWorkout
        .where((e) => e.isDone)
        .fold(0, (sum, e) => sum + e.calories);
    final totalMinutes = doneCount * 5;

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Прогресс за сегодня',
              style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
          const SizedBox(height: 16),
          _buildGoalCard(context,
              emoji: '🎯',
              title: 'Упражнения',
              current: doneCount,
              goal: todayWorkout.length,
              unit: 'шт.',
              color: Colors.teal),
          const SizedBox(height: 12),
          _buildGoalCard(context,
              emoji: '🔥',
              title: 'Калории',
              current: totalCalories,
              goal: calorieGoal,
              unit: 'ккал',
              color: Colors.orange),
          const SizedBox(height: 12),
          _buildGoalCard(context,
              emoji: '⏱',
              title: 'Время',
              current: totalMinutes,
              goal: 45,
              unit: 'мин',
              color: Colors.blue),
          const SizedBox(height: 24),
          const Text('Выполнено:',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          ...todayWorkout.where((e) => e.isDone).map((e) => Padding(
                padding: const EdgeInsets.only(bottom: 4),
                child: Row(
                  children: [
                    Text(e.emoji, style: const TextStyle(fontSize: 20)),
                    const SizedBox(width: 8),
                    Text('${e.name} — ${e.calories} ккал',
                        style: const TextStyle(fontSize: 15)),
                  ],
                ),
              )),
          if (doneCount == 0)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text(
                  'Пока ничего. Отметьте упражнения на вкладке "Тренировка"!',
                  style: TextStyle(
                      color: Colors.grey[500], fontStyle: FontStyle.italic)),
            ),
        ],
      ),
    );
  }

  Widget _buildGoalCard(BuildContext context,
      {required String emoji,
      required String title,
      required int current,
      required int goal,
      required String unit,
      required Color color}) {
    final progress = goal > 0 ? (current / goal).clamp(0.0, 1.0) : 0.0;
    final percent = (progress * 100).toInt();

    return Card(
      clipBehavior: Clip.antiAlias,
      child: Stack(
        children: [
          Container(
            height: 110,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [color.withOpacity(0.15), color.withOpacity(0.05)],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(emoji, style: const TextStyle(fontSize: 28)),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(title,
                          style: const TextStyle(
                              fontSize: 18, fontWeight: FontWeight.bold)),
                    ),
                    Text('$percent%',
                        style: TextStyle(
                            fontSize: 22,
                            fontWeight: FontWeight.bold,
                            color: color)),
                  ],
                ),
                const SizedBox(height: 12),
                ClipRRect(
                  borderRadius: BorderRadius.circular(6),
                  child: LinearProgressIndicator(
                    value: progress,
                    minHeight: 10,
                    backgroundColor: Colors.grey[300],
                    valueColor: AlwaysStoppedAnimation(color),
                  ),
                ),
                const SizedBox(height: 6),
                Text('$current / $goal $unit',
                    style: TextStyle(color: Colors.grey[600], fontSize: 13)),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// ==============================
// Вкладка 3: Настройки
// ==============================
class SettingsTab extends StatefulWidget {
  final VoidCallback onChanged;
  const SettingsTab({super.key, required this.onChanged});

  @override
  State<SettingsTab> createState() => _SettingsTabState();
}

class _SettingsTabState extends State<SettingsTab> {
  bool _notifications = true;
  bool _sound = false;
  bool _vibration = true;

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Настройки',
              style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
          const SizedBox(height: 16),
          Card(
            child: Column(
              children: [
                SwitchListTile(
                  title: const Text('Уведомления'),
                  subtitle: const Text('Напоминания о тренировке'),
                  secondary: const Icon(Icons.notifications),
                  value: _notifications,
                  onChanged: (val) => setState(() => _notifications = val),
                ),
                const Divider(height: 1),
                SwitchListTile(
                  title: const Text('Звук'),
                  subtitle: const Text('Звуковые эффекты'),
                  secondary: const Icon(Icons.volume_up),
                  value: _sound,
                  onChanged: (val) => setState(() => _sound = val),
                ),
                const Divider(height: 1),
                SwitchListTile(
                  title: const Text('Вибрация'),
                  subtitle: const Text('При выполнении упражнения'),
                  secondary: const Icon(Icons.vibration),
                  value: _vibration,
                  onChanged: (val) => setState(() => _vibration = val),
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.local_fire_department,
                            color: Colors.orange),
                        const SizedBox(width: 8),
                        const Text('Цель по калориям',
                            style: TextStyle(
                                fontSize: 18, fontWeight: FontWeight.bold)),
                        const Spacer(),
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 12, vertical: 4),
                          decoration: BoxDecoration(
                              color: Colors.orange.withOpacity(0.15),
                              borderRadius: BorderRadius.circular(20)),
                          child: Text('$calorieGoal ккал',
                              style: const TextStyle(
                                  fontWeight: FontWeight.bold,
                                  color: Colors.orange)),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Slider(
                      value: calorieGoal.toDouble(),
                      min: 100,
                      max: 800,
                      divisions: 14,
                      label: '$calorieGoal ккал',
                      onChanged: (val) {
                        setState(() => calorieGoal = val.toInt());
                        widget.onChanged();
                      },
                    ),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text('100 ккал',
                            style: TextStyle(
                                color: Colors.grey[500], fontSize: 12)),
                        Text('800 ккал',
                            style: TextStyle(
                                color: Colors.grey[500], fontSize: 12)),
                      ],
                    ),
                  ]),
            ),
          ),
        ],
      ),
    );
  }
}

// ==============================
// Вкладка 4: История
// ==============================
class HistoryTab extends StatelessWidget {
  const HistoryTab({super.key});

  @override
  Widget build(BuildContext context) {
    if (history.isEmpty) {
      return const Center(
          child: Text('История пуста. Завершите хотя бы один день!'));
    }

    return ListView.builder(
      padding: const EdgeInsets.all(12),
      itemCount: history.length,
      itemBuilder: (context, index) {
        final day = history[index];
        return Card(
          margin: const EdgeInsets.only(bottom: 8),
          child: ListTile(
            leading: const Icon(Icons.history),
            title: Text('Дата: ${day['date']}'),
            subtitle: Text(
                'Упражнения: ${day['exercises']}, Калории: ${day['calories']}'),
          ),
        );
      },
    );
  }
}

// ==============================
// Экран таймера
// ==============================
class TimerScreen extends StatefulWidget {
  final Exercise exercise;
  const TimerScreen({super.key, required this.exercise});

  @override
  State<TimerScreen> createState() => _TimerScreenState();
}

class _TimerScreenState extends State<TimerScreen> {
  int _seconds = 30; // Таймер на 30 секунд
  Timer? _timer;

  // Старт таймера
  void _startTimer() {
    if (_timer != null && _timer!.isActive) return;
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_seconds > 0) {
        setState(() => _seconds--);
      } else {
        _pauseTimer();
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('Подход завершён!')));
      }
    });
  }

  // Пауза таймера
  void _pauseTimer() {
    _timer?.cancel();
  }

  // Сброс таймера
  void _resetTimer() {
    _pauseTimer();
    setState(() => _seconds = 30);
  }

  String get _formattedTime {
    final minutes = (_seconds ~/ 60).toString().padLeft(2, '0');
    final seconds = (_seconds % 60).toString().padLeft(2, '0');
    return '$minutes:$seconds';
  }

  @override
  void dispose() {
    _pauseTimer();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.exercise.name),
      ),
      body: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(widget.exercise.name,
                style:
                    const TextStyle(fontSize: 28, fontWeight: FontWeight.bold)),
            const SizedBox(height: 16),
            Text('${widget.exercise.sets} подходов',
                style: const TextStyle(fontSize: 20)),
            const SizedBox(height: 32),
            Text(_formattedTime,
                style:
                    const TextStyle(fontSize: 64, fontWeight: FontWeight.bold)),
            const SizedBox(height: 32),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                ElevatedButton(
                    onPressed: _startTimer, child: const Text('Старт')),
                const SizedBox(width: 16),
                ElevatedButton(
                    onPressed: _pauseTimer, child: const Text('Пауза')),
                const SizedBox(width: 16),
                ElevatedButton(
                    onPressed: _resetTimer, child: const Text('Сброс')),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
