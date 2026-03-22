import 'package:flutter/material.dart';

void main() {
  runApp(const CafeApp());
}

class CafeApp extends StatelessWidget {
  const CafeApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Кафе "У Flutter"',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.brown,
        ),
        useMaterial3: true,
      ),
      home: const CategoriesScreen(),
    );
  }
}

// Глобальная переменная для корзины
List<Map<String, dynamic>> cart = [];

// ==============================
// МОДЕЛЬ ДАННЫХ
// ==============================

class Dish {
  final String name;
  final String description;
  final double price;
  final String emoji;
  final List<String> ingredients;

  const Dish({
    required this.name,
    required this.description,
    required this.price,
    required this.emoji,
    required this.ingredients,
  });
}

class Category {
  final String name;
  final String emoji;
  final Color color;
  final List<Dish> dishes;

  const Category({
    required this.name,
    required this.emoji,
    required this.color,
    required this.dishes,
  });
}

// ==============================
// ДАННЫЕ МЕНЮ
// ==============================

final List<Category> menuData = [
  Category(
    name: 'Основные блюда',
    emoji: '🍳',
    color: Colors.green,
    dishes: [
      Dish(
        name: 'Пюре с котлетой',
        description: 'Нежнейшее пюре картошки '
            'с говяжью котлетой.',
        price: 300,
        emoji: '🍗',
        ingredients: ['Картошка', 'Яйца', 'Молоко', 'Фарш', 'Хлеб'],
      ),
      Dish(
        name: 'Свинина с капустой',
        description: 'Свинина, запечённая с капустой и картошкой.',
        price: 500,
        emoji: '🍖',
        ingredients: ['Свинина', 'Капуста', 'Картофель', 'Чеснок', 'Масло'],
      ),
      Dish(
        name: 'Бифштекс с рисом',
        description: 'Сочные бифштексы, поданные с отварным рисом и соусом.',
        price: 600,
        emoji: '🥩',
        ingredients: ['Говядина', 'Рис', 'Лук', 'Масло', 'Соус'],
      ),
    ],
  ),
  Category(
    name: 'Завтраки',
    emoji: '🍳',
    color: Colors.orange,
    dishes: [
      Dish(
        name: 'Омлет с сыром',
        description: 'Пышный омлет из трёх яиц '
            'с тёртым сыром и зеленью.',
        price: 250,
        emoji: '🍳',
        ingredients: ['Яйца', 'Сыр', 'Молоко', 'Зелень'],
      ),
      Dish(
        name: 'Овсянка с ягодами',
        description: 'Овсяная каша на молоке '
            'со свежими ягодами и мёдом.',
        price: 200,
        emoji: '🥣',
        ingredients: ['Овсянка', 'Молоко', 'Ягоды', 'Мёд'],
      ),
      Dish(
        name: 'Блинчики',
        description: 'Тонкие блинчики с вареньем '
            'или сметаной на выбор.',
        price: 220,
        emoji: '🥞',
        ingredients: ['Мука', 'Яйца', 'Молоко', 'Варенье'],
      ),
    ],
  ),
  Category(
    name: 'Супы',
    emoji: '🍲',
    color: Colors.red,
    dishes: [
      Dish(
        name: 'Борщ',
        description: 'Наваристый борщ со сметаной '
            'и чесночными пампушками.',
        price: 320,
        emoji: '🍲',
        ingredients: ['Свёкла', 'Капуста', 'Картофель', 'Мясо', 'Сметана'],
      ),
      Dish(
        name: 'Куриный бульон',
        description: 'Лёгкий куриный бульон '
            'с лапшой и зеленью.',
        price: 280,
        emoji: '🍜',
        ingredients: ['Курица', 'Лапша', 'Морковь', 'Зелень'],
      ),
    ],
  ),
  Category(
    name: 'Напитки',
    emoji: '☕',
    color: Colors.brown,
    dishes: [
      Dish(
        name: 'Капучино',
        description: 'Классический капучино '
            'с молочной пенкой.',
        price: 180,
        emoji: '☕',
        ingredients: ['Эспрессо', 'Молоко'],
      ),
      Dish(
        name: 'Чай с лимоном',
        description: 'Чёрный чай с долькой лимона '
            'и мёдом.',
        price: 120,
        emoji: '🍵',
        ingredients: ['Чай', 'Лимон', 'Мёд'],
      ),
      Dish(
        name: 'Морс',
        description: 'Домашний ягодный морс '
            'из клюквы и брусники.',
        price: 150,
        emoji: '🧃',
        ingredients: ['Клюква', 'Брусника', 'Сахар'],
      ),
    ],
  ),
  Category(
    name: 'Десерты',
    emoji: '🍰',
    color: Colors.pink,
    dishes: [
      Dish(
        name: 'Чизкейк',
        description: 'Нежный чизкейк с ягодным '
            'соусом.',
        price: 350,
        emoji: '🍰',
        ingredients: ['Сливочный сыр', 'Печенье', 'Ягодный соус'],
      ),
      Dish(
        name: 'Тирамису',
        description: 'Итальянский десерт '
            'с маскарпоне и кофе.',
        price: 380,
        emoji: '🍮',
        ingredients: ['Маскарпоне', 'Савоярди', 'Кофе', 'Какао'],
      ),
    ],
  ),
];

// ==============================
// ЭКРАН 1: КАТЕГОРИИ (С ПОИСКОМ)
// ==============================

class CategoriesScreen extends StatefulWidget {
  const CategoriesScreen({super.key});

  @override
  State<CategoriesScreen> createState() => _CategoriesScreenState();
}

class _CategoriesScreenState extends State<CategoriesScreen> {
  String _searchText = '';
  final TextEditingController _searchController = TextEditingController();

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Кафе "У Flutter"'),
        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
        actions: [
          IconButton(
            icon: const Icon(Icons.shopping_cart),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => const CartScreen()),
              );
            },
          ),
        ],
      ),
      body: Column(
        children: [
          // Поле поиска
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: TextField(
              controller: _searchController,
              decoration: InputDecoration(
                hintText: 'Поиск блюд...',
                prefixIcon: const Icon(Icons.search),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                suffixIcon: _searchText.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear),
                        onPressed: () {
                          _searchController.clear();
                          setState(() {
                            _searchText = '';
                          });
                        },
                      )
                    : null,
              ),
              onChanged: (value) {
                setState(() {
                  _searchText = value;
                });
              },
            ),
          ),
          // Контент (категории или результаты поиска)
          Expanded(
            child: _buildContent(),
          ),
        ],
      ),
    );
  }

  Widget _buildContent() {
    if (_searchText.isEmpty) {
      // Показываем категории
      return ListView.builder(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: menuData.length,
        itemBuilder: (context, index) {
          final category = menuData[index];
          return _buildCategoryCard(context, category);
        },
      );
    } else {
      // Фильтруем блюда по всем категориям
      final allDishes = menuData
          .expand((cat) => cat.dishes)
          .where((dish) =>
              dish.name.toLowerCase().contains(_searchText.toLowerCase()))
          .toList();

      if (allDishes.isEmpty) {
        return const Center(child: Text('Ничего не найдено'));
      }

      return ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: allDishes.length,
        itemBuilder: (context, index) {
          final dish = allDishes[index];
          // Находим категорию, к которой относится блюдо (для цвета)
          final category = menuData.firstWhere(
            (cat) => cat.dishes.contains(dish),
          );
          return _buildSearchResultCard(context, dish, category);
        },
      );
    }
  }

  Widget _buildCategoryCard(BuildContext context, Category category) {
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => DishesScreen(category: category),
            ),
          );
        },
        child: Container(
          height: 100,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [
                category.color.withOpacity(0.7),
                category.color.withOpacity(0.3),
              ],
              begin: Alignment.centerLeft,
              end: Alignment.centerRight,
            ),
          ),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            child: Row(
              children: [
                Text(
                  category.emoji,
                  style: const TextStyle(fontSize: 44),
                ),
                const SizedBox(width: 20),
                Expanded(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        category.name,
                        style: const TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '${category.dishes.length} позиций',
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.white.withOpacity(0.8),
                        ),
                      ),
                    ],
                  ),
                ),
                const Icon(
                  Icons.arrow_forward_ios,
                  color: Colors.white70,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSearchResultCard(
      BuildContext context, Dish dish, Category category) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => DishDetailScreen(
                dish: dish,
                categoryColor: category.color,
              ),
            ),
          );
        },
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Container(
                width: 56,
                height: 56,
                decoration: BoxDecoration(
                  color: category.color.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Center(
                  child: Text(
                    dish.emoji,
                    style: const TextStyle(fontSize: 28),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      dish.name,
                      style: const TextStyle(
                        fontSize: 17,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      dish.description,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        fontSize: 13,
                        color: Colors.grey[600],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: category.color.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  '${dish.price.toInt()} ₽',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: category.color,
                    fontSize: 15,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ==============================
// ЭКРАН 2: СПИСОК БЛЮД
// ==============================

class DishesScreen extends StatelessWidget {
  final Category category;

  const DishesScreen({
    super.key,
    required this.category,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('${category.emoji} ${category.name}'),
        backgroundColor: category.color.withOpacity(0.3),
      ),
      body: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: category.dishes.length,
        itemBuilder: (context, index) {
          final dish = category.dishes[index];
          return _buildDishCard(context, dish);
        },
      ),
    );
  }

  Widget _buildDishCard(BuildContext context, Dish dish) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => DishDetailScreen(
                dish: dish,
                categoryColor: category.color,
              ),
            ),
          );
        },
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Container(
                width: 56,
                height: 56,
                decoration: BoxDecoration(
                  color: category.color.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Center(
                  child: Text(
                    dish.emoji,
                    style: const TextStyle(fontSize: 28),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      dish.name,
                      style: const TextStyle(
                        fontSize: 17,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      dish.description,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        fontSize: 13,
                        color: Colors.grey[600],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: category.color.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  '${dish.price.toInt()} ₽',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: category.color,
                    fontSize: 15,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ==============================
// ЭКРАН 3: КАРТОЧКА БЛЮДА (с добавлением в корзину)
// ==============================

class DishDetailScreen extends StatefulWidget {
  final Dish dish;
  final Color categoryColor;

  const DishDetailScreen({
    super.key,
    required this.dish,
    required this.categoryColor,
  });

  @override
  State<DishDetailScreen> createState() => _DishDetailScreenState();
}

class _DishDetailScreenState extends State<DishDetailScreen> {
  int _quantity = 1;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.dish.name),
        backgroundColor: widget.categoryColor.withOpacity(0.3),
      ),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(
              height: 200,
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    widget.categoryColor.withOpacity(0.4),
                    widget.categoryColor.withOpacity(0.1),
                  ],
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                ),
              ),
              child: Center(
                child: Text(
                  widget.dish.emoji,
                  style: const TextStyle(fontSize: 100),
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    widget.dish.name,
                    style: const TextStyle(
                      fontSize: 28,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    '${widget.dish.price.toInt()} ₽',
                    style: TextStyle(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                      color: widget.categoryColor,
                    ),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    widget.dish.description,
                    style: const TextStyle(
                      fontSize: 16,
                      height: 1.5,
                    ),
                  ),
                  const SizedBox(height: 24),
                  const Text(
                    'Состав:',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: widget.dish.ingredients
                        .map(
                          (item) => Chip(
                            avatar: Icon(
                              Icons.check_circle,
                              size: 18,
                              color: widget.categoryColor,
                            ),
                            label: Text(item),
                          ),
                        )
                        .toList(),
                  ),
                  const SizedBox(height: 32),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      IconButton(
                        onPressed: () {
                          if (_quantity > 1) {
                            setState(() {
                              _quantity--;
                            });
                          }
                        },
                        icon: const Icon(Icons.remove_circle_outline),
                        iconSize: 36,
                        color: widget.categoryColor,
                      ),
                      const SizedBox(width: 16),
                      Text(
                        '$_quantity',
                        style: const TextStyle(
                          fontSize: 28,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(width: 16),
                      IconButton(
                        onPressed: () {
                          setState(() {
                            _quantity++;
                          });
                        },
                        icon: const Icon(Icons.add_circle_outline),
                        iconSize: 36,
                        color: widget.categoryColor,
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  SizedBox(
                    width: double.infinity,
                    height: 54,
                    child: ElevatedButton(
                      onPressed: () {
                        // Добавляем товар в корзину
                        cart.add({
                          'name': widget.dish.name,
                          'price': widget.dish.price,
                          'quantity': _quantity,
                          'emoji': widget.dish.emoji,
                        });

                        final total = widget.dish.price * _quantity;
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(
                            content: Text(
                              '${widget.dish.name} x$_quantity = ${total.toInt()} ₽ добавлено в заказ!',
                            ),
                            backgroundColor: widget.categoryColor,
                          ),
                        );
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: widget.categoryColor,
                        foregroundColor: Colors.white,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      child: Text(
                        'В заказ — '
                        '${(widget.dish.price * _quantity).toInt()} ₽',
                        style: const TextStyle(fontSize: 18),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ==============================
// ЭКРАН 4: КОРЗИНА
// ==============================

class CartScreen extends StatelessWidget {
  const CartScreen({super.key});

  @override
  Widget build(BuildContext context) {
    double totalAmount = 0;
    for (var item in cart) {
      totalAmount += item['price'] * item['quantity'];
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Корзина'),
        backgroundColor: Colors.blue,
      ),
      body: cart.isEmpty
          ? const Center(child: Text('Корзина пуста'))
          : ListView.builder(
              itemCount: cart.length,
              itemBuilder: (context, index) {
                final item = cart[index];
                return ListTile(
                  leading:
                      Text(item['emoji'], style: const TextStyle(fontSize: 32)),
                  title: Text(item['name']),
                  subtitle:
                      Text('Цена: ${item['price']} ₽ x ${item['quantity']}'),
                  trailing:
                      Text('${(item['price'] * item['quantity']).toInt()} ₽'),
                );
              },
            ),
      bottomNavigationBar: cart.isEmpty
          ? null
          : Padding(
              padding: const EdgeInsets.all(16.0),
              child: ElevatedButton(
                onPressed: () {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                        content: Text(
                            'Заказ на сумму ${totalAmount.toInt()} ₽ оформлен!')),
                  );
                },
                child: Text('Оформить заказ — ${totalAmount.toInt()} ₽'),
              ),
            ),
    );
  }
}
