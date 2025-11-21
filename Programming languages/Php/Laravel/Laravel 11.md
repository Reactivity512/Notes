# Laravel 11

Laravel 11 вышел 12 марта 2024 года.
Laravel 11 требует, как минимум, версию PHP 8.2.

### Упрощенная структура приложения

В Laravel 11 значительно упростили структуру приложения по умолчанию:
* Удалены лишние файлы, такие как `app/Http/Kernel.php`, `app/Console/Kernel.php`, и `app/Providers/RouteServiceProvider.php`.
* Маршруты (**routes**) теперь регистрируются напрямую в `bootstrap/app.php`.
* Промежуточное (**middleware**) теперь настраивается в `bootstrap/app.php`.

Пример новой структуры:

```php
// bootstrap/app.php
$app = Illuminate\Foundation\Application::configure()
   ->withProviders()
   ->withMiddleware([
       'web' => [
           \App\Http\Middleware\EncryptCookies::class,
           \Illuminate\Cookie\Middleware\AddQueuedCookiesToResponse::class,
           // ...
       ],
   ])
   ->withRouting(
       web: __DIR__.'/../routes/web.php',
       api: __DIR__.'/../routes/api.php',
   )
   ->create();
```

### Healthcheck-маршруты

Laravel 11 включает health-check эндпоинт `/up`, это особенно полезно для интеграции с системами мониторинга, такими как **Kubernetes**, **Healthchecks.io** или другими инструментами, которые проверяют доступность сервисов.
Когда HTTP-запросы отправляются по этому маршруту, Laravel отправляет событие DiagnosingHealth, которое запускает дополнительные проверки. Это может быть проверка подключения к базе данных или работы кэша.
* Если приложение работает корректно, маршрут возвращает HTTP-статус 200 и пустое тело ответа.
* Если приложение не работает (например, база данных недоступна), возвращается HTTP-статус 500.

В контроллер можно добавить проверки, которые вам нужны.

```php
class HealthCheckController extends Controller
{
   public function __invoke(Request $request)
   {
       try {
           // Проверка подключения к базе данных
           DB::connection()->getPdo();

           // Дополнительные проверки (например, Redis)
           // Redis::connection()->ping();

           return response('OK', 200);
       } catch (\Exception $e) {
           return response('Service Unavailable', 500);
       }
   }
}
```

### Улучшения в Eloquent

В Laravel 11 метод `withTrashed()` был улучшен и стал более гибким. Этот метод используется в Eloquent для работы с "мягко удаленными" (soft deleted) записями. Мягкое удаление — это функция, которая позволяет "удалять" записи из базы данных без их физического удаления, помечая их как удаленные с помощью колонки `deleted_at`.

```php
$user->posts()->withTrashed()->get();
```

В Laravel 11 метод `withTrashed()` был улучшен для работы с отношениями (relationships). Теперь его можно использовать для загрузки удаленных записей в связанных моделях.

Пример: предположим, у нас есть модель User, которая имеет отношение hasMany с моделью Post:

```php
class User extends Model
{
   public function posts()
   {
       return $this->hasMany(Post::class);
   }
}

$user = User::find(1);
$posts = $user->posts()->withTrashed()->get(); // Загрузить все посты пользователя, включая удаленные
$deletedPosts = $user->posts()->onlyTrashed()->get(); // Возвращает только удаленные записи.
```

### Улучшения в Eloquent и тестировании

Теперь можно определять касты (casts) прямо в методе `casts()` модели.

```php
class User extends Model
{
    protected function casts(): array
    {
        return [
            'email_verified_at' => 'datetime',
            'options' => 'array',
        ];
    }
}
```

Добавлен новый метод, который упрощает проверку отсутствия ошибок в сессии. Новый метод: `assertSessionHasNoErrors();`
Метод проверяет, что в сессии отсутствуют ошибки валидации. Если ошибки есть, тест завершится с ошибкой.
Laravel 11 теперь поддерживает параллельное тестирование из коробки.

### Улучшения в маршрутизации и очереди

Группировка маршрутов с помощью `Route::group():`
Теперь можно группировать маршруты с помощью замыканий, что делает код более читаемым.

```php
Route::group(function () {
   Route::get('/dashboard', [DashboardController::class, 'index']);
   Route::get('/profile', [ProfileController::class, 'show']);
})->middleware('auth');
```

Новый метод `dispatchAfterResponse():`
Позволяет отправить задачу в очередь после отправки ответа клиенту.

```php
$data = $request->all();
// Логируем данные после отправки ответа
dispatchAfterResponse(function () use ($data) {
   Log::info('Data stored:', $data);
});
return response()->json(['message' => 'Data stored successfully']);
```

### Улучшенная поддержка Dumpable Trait и Новые команды Artisan

Теперь можно использовать метод `dump()` на любом объекте, который использует трейт `Illuminate\Support\Traits\Dumpable`. Это упрощает отладку.

```php
$user = User::first();
$user->dump(); // Выводит информацию о пользователе
```

В Laravel 11 добавили несколько новых команд Artisan:
* `make:class`: Создает класс без шаблона.
* `make:enum`: Создает перечисление (enum).
* `make:interface`: Создает интерфейс.
* `make:trait`: Создает трейт.

### Улучшения в валидации и файловой системе

Новое правило `can`. Проверяет, может ли пользователь выполнить действие.

```php
$request->validate([
   'post_id' => ['required', 'can:update,post'],
]);
```

Новый метод `Storage::json()`. Позволяет читать JSON-файлы напрямую из хранилища.

```php
$data = Storage::json('file.json');
```

### Улучшенная поддержка PHP 8.2 и 8.3, Blade и улучшения в консоли

Laravel 11 полностью поддерживает PHP 8.2 и 8.3, включая новые функции, такие как:
* Анонимные readonly-классы.
* Константы в трейтах.

Улучшения в Blade. Новый синтаксис `@use`. Позволяет импортировать классы прямо в шаблонах Blade.

```php
@use('App\Models\User');
```

Улучшения в консоли. Новый метод `confirmWithTimeout()`. Позволяет задать таймаут для подтверждения действия.

```php
if ($this->confirmWithTimeout('Вы уверены?', 10)) {
    // Действие
}
```
