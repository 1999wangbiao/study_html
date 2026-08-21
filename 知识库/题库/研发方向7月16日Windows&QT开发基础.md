# 研发方向7月16日Windows&QT开发基础

### 1. 阅读以下实现窗口定时器的代码，下列说法中正确的有哪些？（7分）

> ```cpp
> #include <windows.h>
> 
> LRESULT CALLBACK WndProc(HWND hwnd, UINT msg, WPARAM wParam, LPARAM lParam) {
>     switch (msg) {
>         case WM_CREATE: {
>             UINT_PTR timer1 = SetTimer(hwnd, 1, 1000, nullptr);
>             UINT_PTR timer2 = SetTimer(hwnd, 2, 500, nullptr);
>             if (timer1 == 0 || timer2 == 0) {}
>             break;
>         }
>         case WM_TIMER: {
>             if (wParam == 1) {
>                 KillTimer(hwnd, 1);
>                 OutputDebugStringA("Timer 1 triggered and killed\n");
>             }
>             else if (wParam == 2) {
>                 SetTimer(hwnd, 2, 2000, nullptr);
>                 OutputDebugStringA("Timer 2 interval updated\n");
>             }
>             break;
>         }
>         case WM_DESTROY:
>             PostQuitMessage(0);
>             break;
>         default:
>             return DefWindowProc(hwnd, msg, wParam, lParam);
>     }
>     return 0;
> }
> 
> int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE, LPSTR, int nShowCmd) {
>     const wchar_t CLASS_NAME[] = L"Sample Window Class";
>     WNDCLASS wc = {};
>     wc.lpfnWndProc = WndProc;
>     wc.hInstance = hInstance;
>     wc.lpszClassName = CLASS_NAME;
>     RegisterClass(&wc);
>     HWND hwnd = CreateWindowEx(0, CLASS_NAME, L"Sample", WS_OVERLAPPEDWINDOW,
>         CW_USEDEFAULT, CW_USEDEFAULT, CW_USEDEFAULT, CW_USEDEFAULT,
>         nullptr, nullptr, hInstance, nullptr);
>     if (hwnd == nullptr) return 0;
>     ShowWindow(hwnd, nShowCmd);
>     MSG msg;
>     while (GetMessage(&msg, hwnd, 0, 0) > 0) {
>         TranslateMessage(&msg);
>         DispatchMessage(&msg);
>     }
>     return (int)msg.wParam;
> }
> 
> ```

**A.** 第一次 WM_TIMER(wParam=1) 触发后，由于调用了 KillTimer，Timer 1 不会再产生新的 WM_TIMER 消息

**B.** 每次处理 wParam=2 的 WM_TIMER 后，Timer 2 的触发间隔会被更新为 2000ms，并重置计时起点

**C.** WM_TIMER 是低优先级的合成消息，只有当消息队列中没有其他更高优先级消息时才会被 GetMessage 取出

**D.** 若窗口销毁时关联的定时器仍处于活跃状态，系统会自动销毁与该窗口关联的定时器，不会泄漏

**正确答案：** ABCD　|　**我的答案：** ABC　|　❌

**答案详解：** Windows 定时器的核心知识点：定时器的生命周期由 SetTimer/KillTimer 与窗口句柄管理，WM_TIMER 是低优先级的合成消息。- ✅ A：第一次 WM_TIMER(wParam=1) 触发时已调用 KillTimer，该定时器随即被销毁，此后不会再产生新的 WM_TIMER 消息。- ✅ B：每次收到 wParam=2 的 WM_TIMER 后代码会重新 SetTimer，将 Timer 2 的触发间隔更新为 2000ms，并以本次调用为新的计时起点。- ✅ C：WM_TIMER 是低优先级的合成消息，只有当消息队列中没有其他更高优先级消息时，GetMessage 才会为其合成并取出 WM_TIMER。- ✅ D：窗口销毁时系统会清理与该窗口关联的所有定时器，不会造成资源泄漏。

---

### 2. 阅读以下代码，程序在主窗口创建后，通过自定义消息通知窗口更新状态，下列说法正确的有哪些？（7分）

> ```cpp
> #define WM_MY_UPDATE (WM_USER + 100)
> 
> LRESULT CALLBACK WndProc(HWND hWnd, UINT msgId, WPARAM wParam, LPARAM lParam) {
>     if (msgId == WM_MY_UPDATE) {
>         int updateFlag = wParam;
>         // 执行更新逻辑
>         return updateFlag;
>     }
>     else if (msgId == WM_CREATE) {
>         CreateThread(NULL, 0, UpdateThread, hWnd, 0, NULL);
>         return 0;
>     }
>     else if (msgId == WM_DESTROY) {
>         PostQuitMessage(0);
>         return 0;
>     }
>     return DefWindowProc(hWnd, msgId, wParam, lParam);
> }
> 
> DWORD WINAPI UpdateThread(LPVOID lpParam) {
>     HWND hWnd = (HWND)lpParam;
>     // 模拟后台耗时工作
>     Sleep(1000);
>     PostMessage(hWnd, WM_MY_UPDATE, 1, 0);
>     return 0;
> }
> 
> ```

**A.** WM_USER+100是合法的自定义消息范围，可用于应用程序内部自定义消息

**B.** 后台线程通过PostMessage发送自定义消息给UI线程窗口，是线程间通信的合法方式

**C.** 如果这里把PostMessage改成SendMessage，一定会导致程序死锁

**D.** 自定义消息WM_MY_UPDATE必须要在窗口类注册前先定义，否则无法使用

**正确答案：** AB　|　**我的答案：** AB　|　✅

**答案详解：** 自定义消息与跨线程通信的知识点：WM_USER 以上是应用程序自定义消息的合法范围，PostMessage 适合跨线程投递消息。- ✅ A：WM_USER 到 WM_APP-1 之间是保留给应用程序自定义的消息区间，WM_USER+100 是合法用法。- ✅ B：PostMessage 将消息放入目标窗口所在线程的消息队列后立即返回，是后台线程通知 UI 线程的合法方式。- ❌ C：SendMessage 跨线程会阻塞发送线程直到窗口过程处理完毕，但不“一定”死锁；只有目标线程不及时处理消息时才可能长时间阻塞或死锁，说法过于绝对。- ❌ D：自定义消息只是一个数值宏，在源代码任何位置定义均可，与窗口类注册的先后顺序无关，无需在注册前预定义。

---

### 3. 阅读以下程序中发送消息的代码片段，关于SendMessage和PostMessage的区别，下列说法正确的有哪些？
假设hWnd是一个已创建的合法窗口句柄。（7分）

> ```cpp
> // 场景1：发送鼠标左键点击消息
> PostMessage(hWnd, WM_LBUTTONDOWN, MK_LBUTTON, MAKELPARAM(100, 100));
> PostMessage(hWnd, WM_LBUTTONUP, 0, MAKELPARAM(100, 100));
> 
> // 场景2：获取窗口标题文本长度
> int len = SendMessage(hWnd, WM_GETTEXTLENGTH, 0, 0);
> 
> ```

**A.** PostMessage会把消息放到目标窗口所在线程的消息队列，然后立即返回，不等待消息处理

**B.** SendMessage会直接调用目标窗口的窗口过程，等待窗口过程处理完成后才会返回

**C.** PostMessage的返回值表示消息是否成功放入队列，不反映消息处理结果

**D.** SendMessage只能发送给同一进程内的窗口，PostMessage可以发送给跨进程窗口

**正确答案：** ABC　|　**我的答案：** ABC　|　✅

**答案详解：** SendMessage 与 PostMessage 的区别：前者同步直达目标窗口过程并等待返回，后者放入队列后立即返回。- ✅ A：PostMessage 把消息送入目标窗口所在线程的消息队列后立即返回，不等待消息处理。- ✅ B：SendMessage 直接调用（或经系统调度）目标窗口的窗口过程，处理完成后才返回。- ✅ C：PostMessage 的返回值只表示消息是否成功放入队列，并不反映消息的处理结果。- ❌ D：SendMessage 同样可以发送给其他进程的窗口（跨线程/跨进程时会由系统传递并阻塞等待），并非只能发送给同一进程内的窗口。

---

### 4. 阅读以下自定义窗口过程的代码，关于WM_PAINT消息的处理，下列说法正确的有哪些？（7分）

> ```cpp
> LRESULT CALLBACK WndProc(HWND hWnd, UINT msgId, WPARAM wParam, LPARAM lParam) {
>     switch (msgId) {
>         case WM_PAINT: {
>             PAINTSTRUCT ps;
>             HDC hdc = BeginPaint(hWnd, &ps);
>             TextOut(hdc, 20, 20, L"Hello Win32", 10);
>             EndPaint(hWnd, &ps);
>             return 0;
>         }
>         case WM_DESTROY: {
>             PostQuitMessage(0);
>             return 0;
>         }
>         default:
>             break;
>     }
>     return DefWindowProc(hWnd, msgId, wParam, lParam);
> }
> 
> ```

**A.** BeginPaint会自动验证窗口的无效区域，清除WM_PAINT消息的产生原因

**B.** 如果窗口客户区需要重绘，系统会向窗口发送WM_PAINT消息

**C.** 若代码中省略EndPaint调用，也不会影响窗口的消息处理，仅会资源泄漏

**D.** 调用InvalidateRect(hWnd, NULL, TRUE)后，会立即触发WM_PAINT消息被处理

**正确答案：** AB　|　**我的答案：** ABC　|　❌

**答案详解：** WM_PAINT 处理的知识点：有效/无效区域机制、BeginPaint/EndPaint 的作用，以及 WM_PAINT 的低优先级特性。- ✅ A：BeginPaint 获取并验证无效区域，清除 WM_PAINT 产生的原因，使下一次不需要重绘的窗口不再产生 WM_PAINT。- ✅ B：客户区变为无效后，系统即向窗口发送 WM_PAINT 消息请求重绘。- ❌ C：省略 EndPaint 不止是资源占用问题，窗口会一直停留在无效状态、WM_PAINT 反复产生，严重影响消息处理，并非“仅会资源泄漏”。- ❌ D：InvalidateRect 只是把区域标记为无效，WM_PAINT 是低优先级消息，要等消息队列空闲时才被取出处理，不会“立即”触发。

---

### 5. 阅读以下Win32窗口程序的入口与消息循环代码，下列说法中正确的有哪些？（7分）

> ```cpp
> #include <windows.h>
> 
> LRESULT CALLBACK WndProc(HWND, UINT, WPARAM, LPARAM);
> 
> int WINAPI WinMain(HINSTANCE hInst, HINSTANCE, LPSTR, int nShowCmd) {
>     WNDCLASS wc = {CS_HREDRAW | CS_VREDRAW, DefWindowProc, 0, 0, hInst, NULL, NULL, NULL, NULL, L"MyClass"};
>     RegisterClass(&wc);
>     HWND hWnd = CreateWindow(L"MyClass", L"Demo", WS_OVERLAPPEDWINDOW, CW_USEDEFAULT, CW_USEDEFAULT, 600, 400, NULL, NULL, hInst, NULL);
>     ShowWindow(hWnd, nShowCmd);
> 
>     MSG msg;
>     while (GetMessage(&msg, hWnd, 0, 0) > 0) {
>         TranslateMessage(&msg);
>         DispatchMessage(&msg);
>     }
>     return (int)msg.wParam;
> }
> 
> ```

**A.** GetMessage调用会仅获取属于该窗口hWnd的消息

**B.** TranslateMessage的作用是将虚拟按键消息转换为字符消息

**C.** 当GetMessage返回0时，消息循环退出，程序即将终止

**D.** DispatchMessage会直接将消息发送给窗口过程WndProc，不需要经过系统调度

**正确答案：** ABC　|　**我的答案：** BC　|　❌

**答案详解：** Win32 消息循环的知识点：GetMessage 可取指定窗口的消息并报告 WM_QUIT，TranslateMessage 负责按键到字符的转换。- ✅ A：代码中的 GetMessage 指定了窗口句柄 hWnd，只取出属于该窗口的消息，其他窗口的消息留在队列中。- ✅ B：TranslateMessage 把 WM_KEYDOWN/WM_KEYUP 等虚拟按键消息转换为 WM_CHAR 字符消息。- ✅ C：GetMessage 返回 0 表示取到了 WM_QUIT，此时消息循环退出，程序即将终止。- ❌ D：DispatchMessage 负责把消息分派给对应窗口的窗口过程执行，但消息仍需经过系统消息队列和消息循环的调度，并非“不需要经过系统调度”。

---

### 6. 自定义Qt绘图控件的代码如下，关于paintEvent方法的实现，下列说法正确的有哪些？（7分）

> ```cpp
> #include <QWidget>
> #include <QPainter>
> 
> class CustomWidget : public QWidget {
>     Q_OBJECT
> public:
>     explicit CustomWidget(QWidget *parent = nullptr) : QWidget(parent) {}
> protected:
>     void paintEvent(QPaintEvent *event) override {
>         QPainter painter(this);
>         painter.setRenderHint(QPainter::Antialiasing);
>         painter.drawEllipse(QRect(10, 10, 100, 100));
> 
>         QPainter *anotherPainter = new QPainter(this);
>         anotherPainter->drawRect(QRect(20, 20, 80, 80));
> 
>         update();
>     }
> };
> 
> ```

**A.** 在paintEvent中实例化QPainter并绑定到当前widget是正确的自定义绘图做法

**B.** 代码运行后会导致无限递归重绘，最终出现栈溢出或程序崩溃

**C.** 同一QPaintDevice上同时存在两个活跃的QPainter是Qt允许的合法行为

**D.** drawEllipse和drawRect绘制的图形都会显示在控件上

**正确答案：** ABD　|　**我的答案：** AB　|　❌

**答案详解：** Qt 自定义绘图的知识点：paintEvent 内创建 QPainter、同一时刻同一 QPaintDevice 只允许一个 QPainter、绘制内容会显示在控件上。- ✅ A：在 paintEvent 中创建 QPainter 并绑定当前 widget 是标准做法，后续绘制指令都作用于该控件。- ✅ B：场景代码在绘制过程中再次引发了重绘（paintEvent 内部触发了新的重绘），导致 paintEvent 被反复调用、无限递归，最终栈溢出或程序崩溃。- ❌ C：Qt 规定同一 QPaintDevice 上同一时刻只能存在一个活动的 QPainter，同时在两个 QPainter 属于非法用法。- ✅ D：QPainter 已绑定控件，drawEllipse 与 drawRect 绘制的图形都会显示在控件上。

---

### 7. Qt主窗口程序中，有如下与菜单栏相关的代码，下列说法正确的有哪些？（7分）

> ```cpp
> #include <QMainWindow>
> #include <QMenu>
> #include <QAction>
> 
> QMainWindow* createMainWindow() {
>     QMainWindow *mainWin = new QMainWindow;
> 
>     QMenu *fileMenu = mainWin->menuBar()->addMenu("文件(&F)");
>     QAction *openAct = new QAction("打开(&O)", mainWin);
> 
>     fileMenu->addAction(openAct);
>     QObject::connect(openAct, &QAction::triggered, mainWin, [mainWin](){
>         // 处理打开逻辑
>     });
> 
>     return mainWin;
> }
> 
> ```

**A.** menuBar()方法是QMainWindow提供的，默认会在主窗口顶部创建并返回一个菜单栏，无需手动创建

**B.** &F和&O写法用于设置快捷键，按Alt+F可以打开文件菜单，按Alt+O可以触发打开动作

**C.** openAct的父对象是mainWin，当mainWin销毁时openAct会被自动清理

**D.** 必须手动将openAct添加到fileMenu之后，才能将其关联到快捷键，不添加无法使用快捷键

**正确答案：** ABC　|　**我的答案：** ABC　|　✅

**答案详解：** Qt 菜单栏的知识点：menuBar() 由 QMainWindow 自动提供、& 起到 Alt 助记快捷键的作用、QAction 随父对象对象树自动释放。- ✅ A：menuBar() 是 QMainWindow 提供的方法，默认在窗口顶部创建菜单栏并返回，无需手动创建。- ✅ B：&F、&O 将 F、O 指定为助记键，按 Alt+F 可打开文件菜单，按 Alt+O 可触发打开动作。- ✅ C：openAct 加入菜单后被纳入 mainWin 的对象树管理，mainWin 销毁时 openAct 会被自动清理。- ❌ D：快捷键由动作文本中的 & 声明，与是否添加进菜单没有必然关系；加入菜单只是让它可被实际触发，题述“必须添加后才可关联快捷键”错误。

---

### 8. 阅读以下布局管理器代码，运行后窗口大小调整时，下列说法正确的有哪些？（7分）

> ```cpp
> #include <QApplication>
> #include <QWidget>
> #include <QVBoxLayout>
> #include <QPushButton>
> 
> int main(int argc, char *argv[]) {
>     QApplication app(argc, argv);
>     QWidget *window = new QWidget;
>     QVBoxLayout *layout = new QVBoxLayout(window);
> 
>     QPushButton *btn1 = new QPushButton("Button 1");
>     QPushButton *btn2 = new QPushButton("Button 2");
> 
>     layout->addWidget(btn1, 1);
>     layout->addWidget(btn2, 2);
> 
>     window->show();
>     return app.exec();
> }
> 
> ```

**A.** 两个按钮会沿垂直方向排列，占据窗口的整个客户区

**B.** 当窗口高度增加时，btn1和btn2会按照1:2的比例分配新增的高度

**C.** 如果不调用window->setLayout(layout)，布局不会生效，两个按钮都不会显示

**D.** 布局会自动设置btn1和btn2的父对象为window，无需手动设置

**正确答案：** ABD　|　**我的答案：** ABCD　|　❌

**答案详解：** 布局管理器的知识点：QVBoxLayout 垂直排布、setStretchFactor 分配伸缩比例、setLayout 与对象树父子关系的建立。- ✅ A：QVBoxLayout 使两个按钮沿垂直方向排列，并铺满窗口的客户区。- ✅ B：设置了伸缩因子后，窗口高度增加时新增高度会按 1:2 的比例分配给 btn1 与 btn2。- ❌ C：不调用 setLayout 布局确实不会生效，但按钮仍会被显示（通常按默认位置与最小尺寸摆放），并非“两个按钮都不会显示”。- ✅ D：部件加入布局后，布局会自动将它们的父对象重设为 window，无需手动设置父对象。

---

### 9. 阅读以下Qt信号与槽连接代码，在Qt 5及以上版本中，下列说法正确的有哪些？（7分）

> ```cpp
> #include <QApplication>
> #include <QPushButton>
> #include <QDebug>
> 
> class Helper : public QObject {
>     Q_OBJECT
> public slots:
>     void onButtonClicked() {
>         qDebug() << "Button clicked";
>     }
> };
> 
> int main(int argc, char *argv[]) {
>     QApplication app(argc, argv);
>     QPushButton btn;
>     Helper h;
>     QObject::connect(&btn, &QPushButton::clicked, &h, &Helper::onButtonClicked);
>     return app.exec();
> }
> 
> ```

**A.** 该连接使用了Qt 5引入的函数指针语法，编译期会检查信号和槽的签名是否匹配

**B.** 如果点击按钮，该连接会触发Helper::onButtonClicked()的调用，输出"Button clicked"

**C.** 必须将Helper类定义放到头文件中才能通过moc预处理，否则链接会报错

**D.** 如果把槽的签名改成void onButtonClicked(bool checked)，该连接仍然可以通过编译

**正确答案：** ABC　|　**我的答案：** ABC　|　✅

**答案详解：** Qt 5 函数指针信号槽语法的知识点：编译期签名检查、含 Q_OBJECT 的类需要为 moc 处理而定义在头文件、槽参数须与信号匹配。- ✅ A：connect 使用成员函数指针的 Qt 5 新语法，编译期即检查信号与槽的签名是否匹配。- ✅ B：点击按钮发出 clicked 信号，触发 Helper::onButtonClicked() 槽，输出 “Button clicked”。- ✅ C：带 Q_OBJECT 的类要由 moc 预处理，类定义需放在头文件中以便生成元对象代码，否则编译阶段会因无法生成 moc 文件而在链接时报错。- ❌ D：函数指针语法要求信号与槽参数严格匹配；无参 clicked() 无法连接带 bool 参数的槽，即使信号为重载形式，直接取 &QPushButton::clicked 也会因重载产生歧义，根本无法直接通过编译。

---

### 10. 阅读以下Qt Widgets代码，下列说法中正确的有哪些？（7分）

> ```cpp
> #include <QApplication>
> #include <QPushButton>
> 
> int main(int argc, char *argv[]) {
>     QApplication app(argc, argv);
>     QPushButton button("Click me");
>     button.show();
>     return app.exec();
> }
> 
> ```

**A.** 该代码编译运行后会显示一个标题为"Click me"的独立窗口

**B.** 如果把QPushButton的构造改为new QPushButton("Click me")而不手动delete，会造成内存泄漏

**C.** 调用button.show()会将按钮显示到屏幕，该操作将按钮加入Qt事件循环处理体系

**D.** app.exec()会启动Qt的事件循环，直到所有顶层窗口被关闭后才会返回

**正确答案：** ACD　|　**我的答案：** ACD　|　✅

**答案详解：** Qt Widgets 程序的知识点：栈对象/对象树管理控件生命周期、show() 使控件参与事件循环、app.exec() 的运行机制。- ✅ A：按代码构造并 show 后，程序会显示一个标题为 “Click me” 的独立（顶层）窗口。- ❌ B：本题中按钮是栈对象（生命周期由作用域自动管理），无需手动 delete，不存在泄漏问题；题述建立在“改成 new 且不 delete”的假设上，与本题代码的实际管理方式不符。- ✅ C：show() 显示按钮并使其进入 Qt 事件循环的绘制与事件分发体系。- ✅ D：app.exec() 启动 Qt 事件循环，默认在所有顶层窗口关闭（lastWindowClosed）后退出并返回。

---

### 11. 阅读以下QObject继承与事件过滤的代码，程序运行后点击窗口触发事件，下列说法中正确的有哪些？（6分）

> ```cpp
> #include <QObject>
> #include <QWidget>
> #include <QMouseEvent>
> #include <QDebug>
> 
> class Filter : public QObject {
>     Q_OBJECT
> protected:
>     bool eventFilter(QObject* obj, QEvent* ev) override {
>         if(ev->type() == QEvent::MouseButtonPress) {
>             qDebug() << "Filter catches mouse press";
>             return false;
>         }
>         return QObject::eventFilter(obj, ev);
>     }
> };
> 
> class TargetWidget : public QWidget {
>     Q_OBJECT
> public:
>     Filter* filter;
>     explicit TargetWidget(QWidget* parent = nullptr) : QWidget(parent) {
>         filter = new Filter(this);
>         installEventFilter(filter);
>     }
> protected:
>     void mousePressEvent(QMouseEvent* ev) override {
>         qDebug() << "Target gets mouse press";
>         QWidget::mousePressEvent(ev);
>     }
> };
> 
> int main() {
>     TargetWidget w;
>     w.show();
>     return 0;
> }
> 
> ```

**A.** 程序运行后点击窗口，会先输出"Filter catches mouse press"，然后输出"Target gets mouse press"

**B.** 如果Filter的eventFilter返回true，那么TargetWidget的mousePressEvent将不会收到该事件

**C.** 若filter对象在TargetWidget销毁前被手动删除，程序不会出现访问错误

**D.** QObject允许同一个对象安装多个事件过滤器，多个过滤器会按安装顺序倒序处理事件

**正确答案：** AB　|　**我的答案：** ABD　|　❌

**答案详解：** Qt 事件过滤器的知识点：过滤器先于目标部件收到事件、eventFilter 返回 true 可截断事件传递、过滤器对象生命周期必须长于目标对象。- ✅ A：过滤器安装在 TargetWidget 上，事件先进入 Filter::eventFilter，输出 “Filter catches mouse press” 后返回 false，事件继续传给 TargetWidget，再输出 “Target gets mouse press”。- ✅ B：eventFilter 返回 true 表示事件已被处理并拦截，TargetWidget 的 mousePressEvent 将不会收到该事件。- ❌ C：filter 被手动删除但仍是 TargetWidget 的已安装过滤器，之后事件分发会访问已销毁的对象，必然产生访问错误。- ❌ D：同一对象可安装多个事件过滤器，但 Qt 的实际调用次序是“最后安装的过滤器最先被调用”，题述的“按安装顺序倒序处理事件”与 Qt 事件过滤器调度规则不符。

---

### 12. 阅读以下包含Lambda表达式的信号槽连接代码，下列说法中正确的有哪些？（6分）

> ```cpp
> #include <QObject>
> #include <QTimer>
> #include <QDebug>
> 
> int main() {
>     QTimer* timer = new QTimer(nullptr);
>     int counter = 0;
>     QObject::connect(timer, &QTimer::timeout, [&counter]() {
>         counter++;
>         qDebug() << counter;
>     });
>     timer->start(1000);
>     return 0;
> }
> 
> ```

**A.** 该连接使用了Lambda作为槽函数，支持访问作用域内的局部变量counter

**B.** 若希望counter被正确捕获且访问安全，可以将counter改为按值捕获

**C.** 如果timer的父对象被设置为一个存在的QObject，当父对象销毁时timer会自动释放

**D.** Lambda槽连接会自动绑定Lambda的生命周期，当counter销毁后槽不会被调用

**正确答案：** AC　|　**我的答案：** ABC　|　❌

**答案详解：** Lambda 槽与 QObject 生命周期的知识点：按引用捕获局部变量的安全隐患、父对象自动释放子对象、连接并不随捕获对象销毁而自动断开。- ✅ A：Lambda 捕获了作用域内局部变量 counter，作为槽函数可以访问该变量。- ❌ B：改为按值捕获后，槽内使用的是捕获时刻的一份值拷贝，若需要累计更新 counter，修改只作用于拷贝而不会写回原变量，计数功能无法正常工作，并不能“既正确又安全”。- ✅ C：timer 设置父对象后进入对象树，父对象销毁时 timer 会被自动释放。- ❌ D：连接并不会绑定 lambda 所捕获变量的生命周期；counter 销毁后连接依然存在，槽内再访问 counter 是悬垂引用，属于未定义行为，而非“槽不会被调用”。

---

### 13. 阅读以下自定义事件的代码，下列说法中正确的有哪些？（6分）

> ```cpp
> #include <QObject>
> #include <QEvent>
> #include <QCoreApplication>
> 
> class CustomEvent : public QEvent {
> public:
>     static const QEvent::Type customType = static_cast<QEvent::Type>(QEvent::User + 1);
>     int data;
>     CustomEvent(int d) : QEvent(customType), data(d) {}
> };
> 
> class Receiver : public QObject {
> protected:
>     bool event(QEvent* e) override {
>         if(e->type() == CustomEvent::customType) {
>             processData(static_cast<CustomEvent*>(e)->data);
>             return true;
>         }
>         return QObject::event(e);
>     }
> private:
>     void processData(int d) {}
> };
> 
> int main() {
>     Receiver* r = new Receiver;
>     CustomEvent* ev = new CustomEvent(42);
>     QCoreApplication::postEvent(r, ev);
>     return 0;
> }
> 
> ```

**A.** postEvent提交的事件会由Qt事件循环处理，并且会自动释放ev对象

**B.** 自定义事件的类型值必须大于等于QEvent::User，该代码的类型定义符合规则

**C.** 如果改用sendEvent发送事件，事件会被同步处理，并且需要手动释放ev

**D.** 重写event()方法是处理自定义事件的唯一方式，无法通过事件过滤器处理

**正确答案：** ABC　|　**我的答案：** AB　|　❌

**答案详解：** Qt 自定义事件的知识点：postEvent 队列式投递且由 Qt 释放事件对象、sendEvent 同步处理且所有权归调用者、事件类型值须大于等于 QEvent::User。- ✅ A：postEvent 将事件投递到事件队列由 Qt 事件循环处理，处理完毕后 Qt 会自动释放该事件对象。- ✅ B：自定义事件类型值必须大于等于 QEvent::User，代码中的类型定义符合该规则，可避免与系统事件冲突。- ✅ C：sendEvent 在调用线程中同步分发事件，事件对象的所有权仍归调用者，需要手动释放 ev。- ❌ D：除重写 event() 外，还可以通过事件过滤器（installEventFilter）处理自定义事件，重写 event() 并非唯一方式。

---

### 14. 阅读以下信号槽连接代码，下列说法中正确的有哪些？（6分）

> ```cpp
> #include <QObject>
> #include <QPushButton>
> #include <QDebug>
> 
> class Controller : public QObject {
>     Q_OBJECT
> public slots:
>     void onButtonClicked() {
>         qDebug() << "Button clicked";
>     }
> };
> 
> int main() {
>     QPushButton* btn = new QPushButton;
>     Controller* c = new Controller;
>     connect(btn, &QPushButton::clicked, c, &Controller::onButtonClicked);
>     return 0;
> }
> 
> ```

**A.** 该连接使用了Qt5支持的函数指针语法，编译期会检查信号槽参数

**B.** 若将连接语法改为Qt4的SIGNAL/SLOT宏方式，连接也可以正常工作

**C.** 如果此时btn被销毁，c对象会自动被对象树销毁

**D.** 如果c对象先被delete，点击btn不会触发野指针访问错误

**正确答案：** ABD　|　**我的答案：** AB　|　❌

**答案详解：** Qt 信号槽连接与对象生命周期的知识点：Qt 5 函数指针语法编译期检查、与 Qt 4 宏语法兼容、接收对象销毁时连接自动断开。- ✅ A：connect 使用 Qt 5 的函数指针语法，编译期会检查信号与槽的参数是否匹配。- ✅ B：改用 Qt 4 的 SIGNAL/SLOT 宏方式建立连接，也可以正常工作。- ❌ C：btn 与 c 之间没有父子关系，btn 被销毁并不会连带给 c 的自动销毁。- ✅ D：函数指针连接以 c 为上下文对象，c 先被 delete 时该连接会被 Qt 自动断开，之后点击 btn 不会再调用已销毁的 c，不会触发野指针访问错误。

---

### 15. 阅读以下使用QObject的代码，下列说法中正确的有哪些？（6分）

> ```cpp
> #include <QObject>
> 
> class MyWidget : public QObject {
>     Q_OBJECT
> public:
>     explicit MyWidget(QObject* parent = nullptr) : QObject(parent) {}
> };
> 
> int main() {
>     MyWidget* w1 = new MyWidget();
>     MyWidget* w2 = new MyWidget(w1);
>     delete w1;
>     return 0;
> }
> 
> ```

**A.** MyWidget 继承自QObject后拥有对象树所有权机制

**B.** 删除w1后，w2会被w1的析构函数自动释放

**C.** Q_OBJECT 宏为该类启用了 Qt 元对象系统的支持，使其能够使用信号槽、动态属性、qobject_cast 等特性

**D.** QObject本身是可拷贝构造的，支持浅拷贝

**正确答案：** ABC　|　**我的答案：** AC　|　❌

**答案详解：** QObject 对象树与元对象系统的知识点：父子对象自动释放、Q_OBJECT 宏开启元对象特性、QObject 禁止拷贝。- ✅ A：MyWidget 继承 QObject 后即拥有对象树所有权机制，可管理子对象的生命周期。- ✅ B：w1 为 w2 的父对象，delete w1 会在其析构函数中递归释放子对象，w2 被自动 delete。- ✅ C：Q_OBJECT 宏使类接入 Qt 元对象系统，从而支持信号槽、动态属性、qobject_cast 等特性。- ❌ D：QObject 通过 Q_DISABLE_COPY 禁止拷贝构造与拷贝赋值，不存在“可拷贝构造、支持浅拷贝”的说法。

---

