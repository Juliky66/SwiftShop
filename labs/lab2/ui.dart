import 'package:flutter/material.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatefulWidget {
  const MyApp({super.key});

  @override
  State<MyApp> createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  bool _isDark = false;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Три вкладки',
      theme: _isDark ? ThemeData.dark() : ThemeData.light(),
      home: MainScreen(
        onToggleTheme: () {
          setState(() {
            _isDark = !_isDark;
          });
        },
      ),
    );
  }
}

class MainScreen extends StatefulWidget {
  final VoidCallback onToggleTheme;

  const MainScreen({super.key, required this.onToggleTheme});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  int _currentIndex = 0;

  final List<Widget> _screens = [
    const ProfileScreen(),
    const GalleryScreen(),
    ContactsScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Три вкладки'),
        actions: [
          IconButton(
            icon: const Icon(Icons.brightness_6),
            onPressed: widget.onToggleTheme, // Переключение темы
          ),
        ],
      ),
      body: _screens[_currentIndex],
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentIndex,
        onTap: (index) {
          setState(() {
            _currentIndex = index;
          });
        },
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.person),
            label: 'Профиль',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.photo_library),
            label: 'Галерея',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.contacts),
            label: 'Контакты',
          ),
        ],
      ),
    );
  }
}

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Мой профиль'),
      ),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            children: [
              const CircleAvatar(
                radius: 60,
                backgroundColor: Colors.deepPurple,
                child: Text(
                  'ЮТ',
                  style: TextStyle(
                    fontSize: 40,
                    color: Colors.white,
                  ),
                ),
              ),
              const SizedBox(height: 16),
              const Text(
                'Юлия Титкова',
                style: TextStyle(
                  fontSize: 26,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                'Flutter-разработчик',
                style: TextStyle(
                  fontSize: 16,
                  color: Colors.grey[600],
                ),
              ),
              const SizedBox(height: 24),
              // Информационные карточки
              Card(
                child: Column(
                  children: const [
                    ListTile(
                      leading: Icon(Icons.email, color: Colors.blue),
                      title: Text('Email'),
                      subtitle: Text('julia@gmail.com'),
                    ),
                    Divider(),
                    ListTile(
                      leading: Icon(Icons.phone, color: Colors.green),
                      title: Text('Телефон'),
                      subtitle: Text('+7 (983) 456-78-90'),
                    ),
                    Divider(),
                    ListTile(
                      leading: Icon(Icons.location_on, color: Colors.red),
                      title: Text('Город'),
                      subtitle: Text('Новосибирск'),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              // Новая карточка: Университет
              Card(
                child: Column(
                  children: const [
                    ListTile(
                      leading: Icon(Icons.school, color: Colors.orange),
                      title: Text('Университет'),
                      subtitle: Text(
                          'Новосибирский государственный университет экономики и управления'),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              const Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Интересы',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: const [
                  Chip(
                    avatar: Icon(Icons.sports_gymnastics, size: 18),
                    label: Text('Sport'),
                  ),
                  Chip(
                    avatar: Icon(Icons.phone_android, size: 18),
                    label: Text('Mobile Dev'),
                  ),
                  Chip(
                    avatar: Icon(Icons.analytics_outlined, size: 18),
                    label: Text('Analyst'),
                  ),
                  Chip(
                    avatar: Icon(Icons.design_services_outlined, size: 18),
                    label: Text('Design'),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class GalleryScreen extends StatelessWidget {
  const GalleryScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Галерея'),
      ),
      body: GridView.count(
        crossAxisCount: 2,
        children: List.generate(10, (index) {
          return Image.network(
            'https://picsum.photos/200/300?random=$index',
            fit: BoxFit.cover,
          );
        }),
      ),
    );
  }
}

class ContactsScreen extends StatelessWidget {
  // Список с пятью номерами
  final List<String> contacts = [
    '+7 (983) 111-11-11',
    '+7 (983) 222-22-22',
    '+7 (983) 333-33-33',
    '+7 (983) 444-44-44',
    '+7 (983) 555-55-55',
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Контакты'),
      ),
      body: ListView.builder(
        itemCount: contacts.length, // Количество контактов
        itemBuilder: (context, index) {
          return ListTile(
            title: Text('Контакт ${index + 1}'),
            subtitle: Text(
                'Телефон: ${contacts[index]}'), // Отображаем телефон из списка
            leading: const Icon(Icons.account_circle),
          );
        },
      ),
    );
  }
}
