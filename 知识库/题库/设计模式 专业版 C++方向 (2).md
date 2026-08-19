# 设计模式 专业版 C++方向

### 1. 下面关于设计模式的应用场景，说法不合理的是（5分）

**A.** QT的信号槽机制主要用了观察者模式

**B.** 命令模式常用来实现撤销(undo)、重做(redo)机制

**C.** 享元模式常用于避免创建过多相同或相似的对象实例，从而节省内存空间提升系统性能

**D.** 需要分步骤构造一个复杂的文档对象（如设置标题、正文、页脚），应选用工厂模式

**正确答案：** D　|　**我的答案：** D　|　✅

**答案详解：** 本题要求选出“说法不合理”的一项，核心考点是各设计模式的典型应用场景。- ✅ A：Qt 信号槽将发送方与处理方解耦、状态变化时主动通知，是观察者模式的典型应用，说法合理。- ✅ B：命令模式把每个操作封装为命令对象，天然适合实现撤销(undo)、重做(redo)，说法合理。- ✅ C：享元模式通过共享相同或相似对象减少实例数量、节省内存提升性能，说法合理。- ❌ D：分步骤构造复杂文档（标题、正文、页脚）正是建造者（Builder）模式的适用场景，工厂模式用于创建产品对象而不是分步组装，选用工厂模式不合理。

---

### 2. 下面关于单例模式的描述中合理的是（5分）

**A.** 单例模式书写简洁、使用方便，编程中尽可能多地使用单例模式

**B.** 单例模式是为了保证一个类在程序运行期间可以有多个实例，以提高程序的灵活性

**C.** 单例模式存在存储析构顺序不确定、全局状态难管理、线程安全等问题

**D.** 单例模式在单元测试时不会与其它测试用例产生干扰，可方便地进行单元测试

**正确答案：** C　|　**我的答案：** D　|　❌

**答案详解：** 单例模式合理描述的知识点：单例保证类只有一个实例，但存在全局状态、线程安全、析构顺序等问题。- ❌ A：单例会引入全局状态、提高耦合，应谨慎使用，“尽可能多使用”不合理。- ❌ B：单例模式恰恰是保证一个类在程序运行期间只有一个实例，与“可以有多个实例”相反。- ✅ C：单例存在静态存储析构顺序不确定、全局状态难管理、多线程初始化竞态等问题，说法合理。- ❌ D：单例的全局状态会在测试用例之间共享，容易相互干扰，不利于单元测试的隔离。

---

### 3. 如何改进以下代码使其满足开放封闭原则（5分）

> ```
> class Shape 
> {
> public:
>     virtual ~Shape() = default;
> };
> 
> class Circle : public Shape {};
> class Rectangle : public Shape {};
> 
> class Painter
> {
> public:
>     void Draw(vector<Shape*>& shapes)
>     {
>         for (auto shape : shapes)
>         {
>             if (auto circle = dynamic_cast<Circle*>(shape)) 
>                 DrawCircle(circle);
>             else if (auto rect = dynamic_cast<Rectangle*>(shape)) 
>                 DrawRect(rect);
>             // 每新增一种图形需修改此方法
>         }
>     }
> private:
>     void DrawCircle(Circle* circle) { /* 绘制圆形具体实现 */ }
>     void DrawRect(Rectangle* rect) { /* 绘制矩形具体实现 */ }
> };
> 
> ```

**A.** Shape类增加虚函数Draw，每个图形子类重写该虚函数

**B.** 增加一个新的if语句来实现功能

**C.** 增加ShapeType枚举标识类型，再根据类型调用不同的绘制方法

**D.** 在Painter类中添加一个成员变量来存储图形类型和绘制函数的映射，然后在Draw方法中通过该映射来调用绘制函数

**正确答案：** A　|　**我的答案：** A　|　✅

**答案详解：** 开放封闭原则改进知识点：对扩展开放、对修改封闭，应通过虚函数多态扩展新类型而非修改原判断逻辑。- ✅ A：在 Shape 基类增加虚函数 Draw、各图形子类重写，新增图形只需新增子类，不改动已有类，符合开放封闭原则。- ❌ B：增加新的 if 语句直接改动原有判断代码，属于对已有实现的修改，违背开放封闭原则。- ❌ C：增加 ShapeType 枚举再按类型分派调用，新增图形仍要修改分派代码，未做到对扩展开放。- ❌ D：在 Painter 中维护类型与绘制函数映射，新增图形仍需改动 Painter 的注册/成员逻辑，类型分派责任被耦合进 Painter，扩展性不如虚函数多态。

---

### 4. 结合优化前后的两段代码，请分析优化后的代码主要体现了哪个设计原则？（5分）

> 优化前代码
> 
> 
> ```
> class ReportGenerator
> {
> public:
> 	enum class ExportType
> 	{
> 		JSON,
> 		XML,
> 	};
> 	void generaterReport(ExportType type)
> 	{
> 		switch (type)
> 		{
> 		case JSON:
> 			exportToJSON();
> 			break;
> 		case XML:
> 			exportToXML();
> 			break;
> 		}
> 	}
> private:
> 	void exportToJSON()
> 	{}
> 	void exportToXML()
> 	{}
> };
> 
> 
> ```
> 优化后代码
> 
> 
> ```
> class IDataExport
> {
> public:
> 	virtual void exportData() = 0;
> };
> 
> class JSONExport : public IDataExport
> {
> public:
> 	void exportData() override
> 	{}
> };
> 
> class XMLExport : public IDataExport
> {
> public:
> 	void exportData() override
> 	{}
> };
> 
> 
> class ReportGenerator
> {
> public:
> 	ReportGenerator(IDataExport *ex)
> 		:m_pExporter(ex)
> 	{}
> 	void generateReport()
> 	{}
> private:
> 	IDataExport* m_pExporter;
> 
> };
> 
> ```

**A.** 体现了里氏替换原则，即任何父类对象出现的地方都可以用子类对象来替换，并且替换之后不会影响程序的正确性

**B.** 体现了开放封闭原则。即对扩展开放，对修改封闭。具体指添加新需求时，尽量不去改动已有的类接口及实现，只需增加新功能的代码即可

**C.** 体现了接口隔离原则，即客户端不应该依赖它不需要的接口，一个类对另一个类的依赖应该建立在最小的接口上

**D.** 体现了迪米特法则，即一个类不应该知道另一个类太多的细节，只需要与直接相关的对象进行交互,降低耦合

**正确答案：** A　|　**我的答案：** B　|　❌

**答案详解：** 优化后代码体现设计原则的判断知识点：以基类接口统一操作具体子类对象、子类替换父类后行为保持一致，体现里氏替换原则。- ✅ A：优化后的代码通过基类抽象让具体图形对象可替换基类位置，且替换后程序行为正确，主要体现的正是里氏替换原则。- ❌ B：开放封闭原则侧重“添加新需求时不断言旧代码”，而本题优化强调的是子类对象替换父类并被统一调用，直接体现是里氏替换而非开放封闭。- ❌ C：接口隔离原则强调客户端只依赖最小必要接口，与本题父子替换的场景不直接相关。- ❌ D：迪米特法则强调对象间减少不必要的细节依赖，也与本题“替换”的语义不符。

---

### 5. 观察者模式是一种行为型设计模式，下面关于观察者模式的说法最合理的是?（5分）

**A.** 观察者模式的主要缺点是会增加系统的复杂性，但不会引入额外的性能开销

**B.** 当需要为对象添加新的功能，而不修改其原有代码时，可以考虑选用观察者模式

**C.** 主题(Subject)需要管理观察者(Observer)的注册和注销，若主题在通知观察者时直接传递数据，则需双方约定数据格式（如参数类型）

**D.** 多个观察者之间一般会相互通信，以确保状态更新的一致性

**正确答案：** C　|　**我的答案：** C　|　✅

**答案详解：** 观察者模式合理说法的知识点：主题负责观察者的注册/注销并主动通知、直接传数据需约定格式、观察者之间一般不直接通信。- ❌ A：观察者模式在通知遍历、注册注销上会带来一定开销，“不会引入额外性能开销”说法错误。- ❌ B：“不修改原有代码为对象添加新功能”是装饰器模式（或开闭原则）的典型场景，不是观察者模式的主要用途。- ✅ C：主题需要管理观察者的注册和注销，若通知时直接传递数据，就必须与观察者约定数据格式（如参数类型、字段含义），说法合理。- ❌ D：多个观察者之间通常不直接相互通信，而是各自与主题交互，由主题统一通知保证一致性。

---

### 6. 如何改进型以下代码使其满足开放封闭原则（5分）

> ```
> class Shape
> {
> public:
> 	virtual ~Shape() = default;
> };
> 
> class Circle : public Shape{};
> class Rectangle : public Shape{};
> 
> class Painter
> {
> public:
> 	void draw(std::vector<Shape*> &shapes)
> 	{
> 		for (auto shape : shapes)
> 		{
> 			if (auto circle = dynamic_cast<Circle*>(shape))
> 				drawCircle(circle);
> 			else if (auto rect = dynamic_cast<Rectangle*>(shape))
> 				drawRect(rect);
> //每新增一种图形需修改此方法
> 		}
> 	}
> private:
> 	void drawCircle(Circle* circle){/*绘制圆形的具体实现*/ }
> 	void drawRect(Rectangle *rect){/*绘制矩形的具体实现*/ }
> };
> 
> ```

**A.** Shape类增加虚函数Draw，每个图形子类重写该虚函数

**B.** 增加一个新的if语句来实现功能

**C.** 增加ShapeType枚举标识类型，再根据类型调用不同的绘制方法

**D.** 在Painter类中添加一个成员变量来存储图形类型和绘制函数的映射，然后在Draw方法中通过该映射来调用绘制函数

**正确答案：** A　|　**我的答案：** A　|　✅

**答案详解：** 开放封闭原则改进知识点：扩展新图形应通过虚函数多态，让新类型自然接入，而不是修改已有分派逻辑。- ✅ A：Shape 增加虚函数 Draw、每个图形子类重写，新增图形新增子类即可，不改动已有类，符合开放封闭原则。- ❌ B：增加新的 if 语句是对已有代码的改动，违背“对修改封闭”。- ❌ C：增加枚举再按类型判断调用，新类型仍需修改分派代码，未做到对扩展开放。- ❌ D：在 Painter 中维护类型与绘制函数映射，新增图形仍需修改 Painter 的相关逻辑，扩展性明显差于虚函数多态。

---

### 7. 下面关于单例模式的描述中合理的是?（5分）

**A.** 单例模式书写简洁、使用方便，编程中尽可能多地使用单例模式

**B.** 单例模式是为了保证一个类在程序运行期间可以有多个实例，以提高程序的灵活性

**C.** 单例模式存在存储析构顺序不确定、全局状态难管理、线程安全等问题

**D.** 单例模式在单元测试时不会与其它测试用例产生干扰，可方便地进行单元测试

**正确答案：** C　|　**我的答案：** C　|　✅

**答案详解：** 单例模式合理描述的知识点：唯一实例与随之而来的全局状态、线程安全、析构顺序等代价。- ❌ A：单例引入全局状态和耦合，应谨慎使用，“尽可能多使用”不合理。- ❌ B：单例模式保证一个类只有一个实例，与题意相反。- ✅ C：单例存在存储析构顺序不确定、全局状态难管理、线程安全等问题，说法合理。- ❌ D：单例的全局状态会在测试用例间共享而产生干扰，并非“不会与其它测试用例产生干扰”。

---

### 8. 在下面代码中，Subject::notify()函数的主要作用是（ ）（5分）

> ```
> class Observer
>  {
> public:
>     virtual void update() = 0;
> };
> 
> class Subject
> {
> private:
>     std::vector<Observer*> m_observers;
> public:
>     void attach(Observer* obs)
>     {
>         m_observers.push_back(obs);
>     }
>     void notify()
>     {
>         for (auto obs : m_observers)
>         {
>             obs->update();
>         }
>     }
> };
> 
> class ConcreteObserver : public Observer
> {
> public:
>     void update() override
>     {
>         std::cout << "Received update notification." << std::endl;
>     }
> };
> 
> ```

**A.** 检查是否有新的观察者加入

**B.** 遍历观察者列表，调用每个观察者的update方法，通知它们状态已改变

**C.** 更新被观察者自身的状态

**D.** 从观察者列表中删除无效的观察者

**正确答案：** B　|　**我的答案：** B　|　✅

**答案详解：** 观察者模式中 notify() 职责的知识点：通知阶段遍历观察者列表并调用其 update 方法。- ❌ A：检查是否有新观察者加入是注册（register/attach）操作的职责，与 notify 无关。- ✅ B：notify() 遍历观察者列表，逐一对每个观察者调用 update 方法，通知它们状态已改变，这正是其主要作用。- ❌ C：更新被观察者自身的状态是 setState/setData 等方法的职责。- ❌ D：从观察者列表中删除无效观察者属于注销（detach/remove）操作的职责。

---

### 9. 在工厂模式中，产品类的抽象基类的主要作用是（5分）

**A.** 为了让产品类的代码更加复杂

**B.** 用于定义产品类的公共接口，使得工厂类可以以统一的方式创建和操作不同的产品

**C.** 只是一种编程习惯，没有实际作用

**D.** 用于限制产品类的数量

**正确答案：** B　|　**我的答案：** B　|　✅

**答案详解：** 工厂模式中产品抽象基类作用的知识点：抽象基类定义产品公共接口，使工厂能以统一方式创建和操作不同的产品。- ❌ A：抽象基类用于抽象公共行为、支持多态，而不是为了把代码变复杂。- ✅ B：产品抽象基类定义公共接口，工厂返回抽象类型、客户端统一操作各类产品，正是其主要作用。- ❌ C：抽象基类是工厂模式实现多态的关键环节，并非无用的编程习惯。- ❌ D：抽象基类不限制产品数量，产品的扩展恰恰通过新增子类来实现。

---

### 10. 合成复用原则提倡（5分）

**A.** 尽量使用继承来达到复用的目的

**B.** 尽量使用对象组合 / 聚合，而不是继承来达到复用的目的

**C.** 复用代码时可以随意选择继承或组合

**D.** 不提倡复用代码，应该尽量编写新的代码

**正确答案：** B　|　**我的答案：** B　|　✅

**答案详解：** 合成复用原则的知识点：优先使用对象组合/聚合而非继承达到复用目的，以降低耦合。- ❌ A：继承会提高类间耦合、破坏封装，合成复用原则并不提倡优先使用继承。- ✅ B：尽量使用对象组合/聚合而不是继承达到复用目的，是合成复用原则的核心主张。- ❌ C：应优先考虑组合而非随意选择，该原则对复用方式有明确倾向。- ❌ D：合成复用原则倡导的是合理复用，并非不提倡复用代码。

---

### 11. 图片切换如何实现撤销重做？（5分）

> 业务场景
> 图片查看器支持 Ctrl+Z 撤销最近一次切换。按钮若直接调 viewer.show(url)，撤销栈会散落各事件处理器；把历史塞进 viewer 又使其职责过重。要求操作可撤销，UI 只发出「执行某操作」的请求。
> 设计思想
> 命令（Command） 将操作封装为带 execute/undo 的对象，由 History 栈统一调度。
> 
> ```js
> class ShowImageCommand {
>   constructor(viewer, url) {
>     this.viewer = viewer;
>     this.url = url;
>     this.prev = null;
>   }
>   execute() {
>     this.prev = this.viewer.current;
>     this.viewer.show(this.url);
>   }
>   undo() {
>     this.viewer.show(this.prev);
>   }
> }
> 
> class History {
>   constructor() { this.stack = []; }
>   run(cmd) { cmd.execute(); this.stack.push(cmd); }
>   undo() {
>     const cmd = this.stack.pop();
>     cmd.undo();
>   }
> }
> 
> const history = new History();
> history.run(new ShowImageCommand(viewer, '/photos/a.jpg'));
> history.run(new ShowImageCommand(viewer, '/photos/b.jpg'));
> history.undo(); // 撤销到 a.jpg
> 
> 
> ```
> 阅读代码后，请选择错误的选项。

**A.** `History` 只依赖命令的 `execute/undo` 接口，不必了解 `viewer` 内部 DOM 细节。

**B.** 命令对象可附带时间戳、操作者等元数据，便于扩展操作审计日志。

**C.** 宏命令可将多个 `execute` 组合为一个命令，支持一次 `undo` 撤销多步操作。

**D.** 为提高性能，`History.undo` 应直接 `viewer.show(cmd.prev)`，跳过 `cmd.undo()` 调用，效果相同。

**正确答案：** D　|　**我的答案：** D　|　✅

**答案详解：** 命令模式实现撤销/重做场景的知识点：历史记录只依赖命令的 execute/undo 接口、命令可携带元数据、宏命令组合多步操作。- ✅ A：History 只依赖命令的 execute/undo 接口，不必了解 viewer 内部 DOM 细节，符合命令模式的解耦思想。- ✅ B：命令对象可附带时间戳、操作者等元数据，便于扩展操作审计日志，说法合理。- ✅ C：宏命令将多个 execute 组合为一个命令，一次 undo 即可撤销多步操作，说法合理。- ❌ D：直接调用 viewer.show(cmd.prev) 跳过了命令封装的 undo 逻辑，命令可能携带的副作用与内部状态无法被正确恢复，且绕过了命令接口、破坏了模式封装，“效果相同”的说法错误。

---

### 12. 栏目树如何统一操作部分与整体？（5分）

> 业务场景
> 新闻 CMS 栏目为树形：「科技」下可有子栏目「AI」，也可直接挂文章。运营需一键统计栏目下文章总数（含子栏目）、删除栏目时连带删子树。早期对文章和栏目分别写删除逻辑，常误删或漏删。
> 设计思想
> 组合（Composite） 叶子与容器实现相同接口（如 count/remove），客户端一致对待部分与整体。
> 
> ```js
> class NewsItem {
>   count() { return 1; }
>   remove() { /* 删文章 */ }
> }
> 
> class NewsCategory {
>   constructor() { this.children = []; }
>   add(node) { this.children.push(node); }
>   count() {
>     return this.children.reduce((sum, n) => sum + n.count(), 0);
>   }
>   remove() {
>     this.children.forEach(n => n.remove());
>   }
> }
> 
> const ai = new NewsCategory('AI');
> ai.add(new NewsItem('AI 专题稿'));
> 
> const root = new NewsCategory('科技');
> root.add(ai);
> root.add(new NewsItem('发布稿'));
> 
> root.count();  // 2，客户端无需区分叶子/容器
> root.remove(); // 级联删除整棵子树
> 
> 
> ```
> 阅读代码后，请选择错误的选项。

**A.** 对 `root.count()` 无需判断节点类型，因为容器与叶子实现了相同的 `count` 接口。

**B.** `NewsCategory.remove` 递归调用子节点 `remove`，体现组合结构的级联操作。

**C.** 组合模式要求容器和叶子必须继承同一抽象基类，否则 `reduce` 递归无法通过类型检查。

**D.** 新增「专题页」节点类型时，只需实现 `count/remove` 同样接口，即可挂入现有树结构。

**正确答案：** C　|　**我的答案：** C　|　✅

**答案详解：** 组合模式统一部分与整体的知识点：容器与叶子实现统一接口实现递归、remove 级联操作、新增节点类型通过实现同一接口接入。- ✅ A：容器与叶子实现相同的 count 接口，对 root.count() 无需判断节点类型即可递归计算，说法合理。- ✅ B：NewsCategory.remove 递归调用子节点的 remove，体现组合结构的级联清理操作，说法合理。- ❌ C：“必须继承同一抽象基类”过于绝对：只要容器与叶子实现同一接口（纯虚基类/接口约定）即可被统一递归处理，类型检查依据的是接口而非某一个具体抽象基类，说法错误。- ✅ D：新增“专题页”节点类型时只需实现 count/remove 相同接口，即可挂入现有树结构，体现组合模式的扩展性，说法合理。

---

### 13. 复杂简历如何分步组装？（5分）

> 业务场景
> 在线简历编辑器有 20+ 字段，教育、项目、技能等模块可选可配，组合差异大。同一份数据需导出 HTML 预览和 JSON 对接 API。超长构造函数曾导致参数顺序错误；新增「实习经历」时所有调用方都要改签名。
> 设计思想
> 建造者（Builder） 分步组装复杂对象，将构建过程与最终表示分离，避免巨型构造函数。
> 
> ```js
> class ResumeBuilder {
>   constructor() {
>     this.parts = { basic: {}, modules: [] };
>   }
>   setBasic(info) {
>     this.parts.basic = info;
>     return this; // 链式调用
>   }
>   addEducation(edu) {
>     this.parts.modules.push({ type: 'edu', data: edu });
>     return this;
>   }
>   addProject(proj) {
>     this.parts.modules.push({ type: 'proj', data: proj });
>     return this;
>   }
>   build(format) {
>     // 组装过程与最终产物分离
>     return format === 'html'
>       ? new HtmlResume(this.parts)
>       : new JsonResume(this.parts);
>   }
> }
> 
> const resume = new ResumeBuilder()
>   .setBasic({ name: 'Li' })
>   .addEducation({ school: 'X' })
>   .addProject({ repo: 'Y' })
>   .build('html');
> 
> 
> ```
> 阅读代码后，请选择错误的选项。

**A.** 链式调用让「缺哪些模块」一目了然，比 20 个位置的构造函数可读性更好。

**B.** `build('html')` 与 `build('json')` 可在最后一步切换产物类型，组装过程可复用。

**C.** 建造者要求所有模块在 `build` 前必须全部填写，否则应抛出异常，这是模式强制约束。

**D.** 新增「实习经历」时，通常只需加 `addInternship` 方法，而不必改所有调用方签名。

**正确答案：** C　|　**我的答案：** C　|　✅

**答案详解：** 建造者模式分步组装的知识点：链式调用提升可读性、组装过程可复用、模块并非强制必填、扩展新模块一般不改调用方签名。- ✅ A：链式调用让“缺哪些模块”一目了然，比 20 个参数的构造函数可读性更好，说法合理。- ✅ B：通过 build 参数在最后一步切换产物类型，组装过程得到复用，说法合理。- ❌ C：建造者模式并不强制要求“所有模块在 build 前必须全部填写”，模块可以有默认值、可以按需跳过；是否校验必填是业务/实现选择，并非模式的强制约束。- ✅ D：新增“实习经历”通常只需增加 addInternship 方法，已有调用方签名不变，扩展成本低，说法合理。

---

### 14. 浅色/深色主题如何成套切换？（5分）

> 业务场景
> 桌面应用主题商店支持浅色/深色一键切换。表单页含按钮、输入框、弹窗，设计规范要求同主题下风格必须一致，不能浅色按钮配深色弹窗。早期为各控件单独建工厂，曾出现 LightButtonFactory 与 DarkDialogFactory 混用导致验收打回。
> 设计思想
> 抽象工厂（Abstract Factory） 用一个工厂接口创建一族相关产品，保证族内风格一致；切换工厂即切换整套 UI。
> 
> ```js
> function DarkThemeFactory() {
>   return {
>     createButton()  { return new DarkButton(); },
>     createInput()   { return new DarkInput(); },
>     createDialog()  { return new DarkDialog(); }
>   };
> }
> 
> function LightThemeFactory() {
>   return {
>     createButton()  { return new LightButton(); },
>     createInput()   { return new LightInput(); },
>     createDialog()  { return new LightDialog(); }
>   };
> }
> 
> function renderForm(factory) {
>   const btn = factory.createButton();
>   const input = factory.createInput();
>   const dialog = factory.createDialog();
>   return { btn, input, dialog };
> }
> 
> // 切换主题 = 切换工厂，保证整套控件同族
> const lightForm = renderForm(LightThemeFactory());
> const darkForm = renderForm(DarkThemeFactory());
> 
> 
> ```
> 阅读代码后，请选择错误的选项。

**A.** 为按钮、输入框、弹窗各建独立工厂更灵活，抽象工厂反而限制了自由组合浅色与深色控件。

**B.** 传入同一个 `factory` 创建整套控件，可避免「浅色按钮 + 深色弹窗」的混搭。

**C.** `renderForm` 只依赖工厂的创建接口，切换主题时替换传入的 factory 即可。

**D.** 若新增 `Tooltip` 控件，所有主题工厂都要加 `createTooltip`，这是该结构的已知代价。

**正确答案：** A　|　**我的答案：** A　|　✅

**答案详解：** 抽象工厂成套切换的知识点：同一工厂保证产品族风格一致、替换工厂实现主题切换、新增产品类型是所有主题工厂的固有扩展代价。- ❌ A：抽象工厂正是为了保证同一产品族（成套控件）风格一致而设计，为每种控件各建工厂反而会造成“浅色按钮+深色弹窗”的混搭；“抽象工厂限制了自由组合”的说法不成立。- ✅ B：用同一个 factory 创建整套控件，可避免浅色按钮与深色弹窗这类混搭，说法合理。- ✅ C：renderForm 只依赖工厂的创建接口，切换主题时替换传入的 factory 即可，说法合理。- ✅ D：新增 Tooltip 控件属于新增产品族成员，所有主题工厂都要增加 createTooltip，这是抽象工厂结构固有的扩展代价，说法合理。

---

### 15. 新增支付渠道如何不改主流程？（5分）

> 业务场景
> 公司内部支付 SDK 的 PaymentService.pay() 被各业务线共用：创建网关 → 扣款 → 返回结果。渠道每季度可能新增。早期在 PaymentService 内用 if/else 选渠道，曾发生 merge 冲突和漏改导致的线上故障。现约定：新增渠道不得修改 PaymentService 主流程。
> 设计思想
> 工厂方法（Factory Method） 将实例化下沉到具体工厂，高层只依赖「能 create 产品」的抽象接口，实现开闭原则。
> 
> ```js
> class PaymentService {
>   pay(order, channelFactory) {
>     const gateway = channelFactory.createGateway();
>     return gateway.charge(order);
>   }
> }
> 
> class WechatFactory {
>   createGateway() { return new WechatGateway(); }
> }
> 
> class AlipayFactory {
>   createGateway() { return new AlipayGateway(); }
> }
> 
> // 调用方注入具体工厂，主流程不感知渠道类名
> const service = new PaymentService();
> service.pay(order, new WechatFactory());
> service.pay(order, new AlipayFactory());
> 
> ```

**A.** 新增银联渠道时，可只增加 `UnionPayFactory`，而不修改 `PaymentService.pay` 方法体。

**B.** `PaymentService` 依赖的是 `createGateway` 能力，而非具体渠道类名，符合依赖倒置。

**C.** 若把 `switch(channel)` 写回 `PaymentService` 内部，每增渠道都要改主流程，违背开闭原则。

**D.** 工厂方法主要用于替换 `charge` 扣款算法；只要把算法表化，就无需工厂类。

**正确答案：** D　|　**我的答案：** D　|　✅

**答案详解：** 工厂方法与依赖倒置的知识点：新增渠道通过新增工厂类而不改动主流程、客户端依赖抽象创建能力而非具体类名、策略替换算法不等于工厂职责。- ✅ A：新增银联渠道只需增加 UnionPayFactory，不必修改 PaymentService.pay 的方法体，符合开闭原则。- ✅ B：PaymentService 依赖的是 createGateway 这一抽象能力，而非具体渠道类名，符合依赖倒置原则。- ❌ C：把 switch(channel) 写回 PaymentService 内部后，每新增渠道都要修改主流程，违背开闭原则。- ❌ D：工厂方法的核心职责是封装创建渠道对象的逻辑，而“替换 charge 扣款算法”是策略模式的职责；仅把算法表化并不能代替工厂对对象创建的封装，说法错误。

---

### 16. 图表组件直接轮询 DataCenter 拉数据，改为观察者订阅后，主要改善了什么？（5分）

> 背景
> 初版图表组件每 500ms 轮询数据中心检查数据是否变化；重构后图表注册为 DataCenter 的观察者，数据更新时主动 notify。
> 对比摘要
> 方式
> 数据更新时
> 耦合
> 轮询
> 定时请求，可能空跑
> 图表需知道如何取数、轮询间隔
> 观察者
> 主题主动推送
> 图表只实现 update 接口
> 改为观察者模式后，主要改善了？

**A.** 图表必须知道 DataCenter 内部所有字段才能工作

**B.** 实现主题与观察者松耦合，数据变化时主动通知，避免无效轮询

**C.** 观察者模式禁止使用 addObserver 和 removeObserver

**D.** 轮询方式比观察者更省资源，应优先使用

**正确答案：** B　|　**我的答案：** B　|　✅

**答案详解：** 观察者模式替代轮询的知识点：主题与观察者松耦合、数据变化主动通知，避免无效轮询浪费资源。- ❌ A：改为订阅后，图表只需依赖通知/数据接口，不必知道 DataCenter 的内部字段，说法错误。- ✅ B：实现主题与观察者松耦合，数据变化时主动通知，避免无效轮询，正是这种改造带来的主要改善。- ❌ C：观察者模式正是通过 addObserver/removeObserver 管理订阅关系，“禁止使用”说法错误。- ❌ D：轮询存在周期空转、实时性差、浪费资源等问题，并不比观察者更省资源。

---

### 17. DataSubject 通知观察者更新数据，下列说法错误的是？（5分）

> 背景
> 实时监控系统用观察者模式：数据变化时自动更新图表和表格组件。
> 相关代码（简化）
> 
> ```js
> class DataSubject {
>   constructor() { this.observers = []; this.data = null; }
>   addObserver(observer) {
>     if (!this.observers.includes(observer)) this.observers.push(observer);
>   }
>   removeObserver(observer) {
>     const i = this.observers.indexOf(observer);
>     if (i !== -1) this.observers.splice(i, 1);
>   }
>   notifyObservers() {
>     this.observers.forEach(o => o.update(this.data));
>   }
>   setData(newData) { this.data = newData; this.notifyObservers(); }
> }
> 
> // lineChart 被 remove 后，再次 setData 只有 barChart 和 dataTable 收到通知
> 
> 
> ```
> 下列说法 错误 的是？

**A.** DataSubject 是主题（Subject），ChartObserver/TableObserver 是观察者（Observer）

**B.** lineChart 被 remove 后，后续 setData 不会再通知 lineChart

**C.** 该实现允许同一观察者被添加多次，可能导致一次更新触发多次 update

**D.** 主题不需要知道观察者的具体类型，只需调用 update 接口，实现了松耦合

**正确答案：** C　|　**我的答案：** C　|　✅

**答案详解：** 观察者模式场景判定的知识点：主题与观察者的角色划分、移除后不再通知、观察者去重、主题对观察者的松耦合。- ✅ A：DataSubject 是主题（Subject），ChartObserver/TableObserver 是观察者（Observer），角色划分正确。- ✅ B：lineChart 被 remove 后即从观察者列表移除，后续 setData 不会再通知 lineChart，说法合理。- ❌ C：维护观察者容器时应保证同一观察者只注册一次（如使用集合去重），不会因重复添加而触发多次 update；若容器不去重则属于实现缺陷而非模式本意，C 的说法不成立。- ✅ D：主题只需调用观察者的 update 接口，无需知道其具体类型，实现了主题与观察者的松耦合，说法合理。

---

### 18. 如何保证全局配置只有唯一入口？（5分）

> 业务场景
> 某图片搜索前端中，「缩略图加载」「详情预取」「失败重试」三个模块都要读取同一套网络策略：maxConcurrent、retryPolicy 等。目前各模块各自 new 配置副本，线上曾出现 A 模块并发 4、B 模块并发 8 的不一致；复制粘贴还带出过字段遗漏。团队要求全局只保留一份可动态调整的运行时配置，且禁止业务方直接 new。
> 设计思想
> 单例（Singleton） 保证进程中只有一个实例，并提供全局访问点。JavaScript 常借助 IIFE 闭包隐藏 instance。适用全局配置、弹窗管理器等——代价是全局可变状态，需约束谁有权修改。
> 
> ```js
> const RequestCenter = (function () {
>   let instance = null;
>   function create() {
>     return {
>       maxConcurrent: 4,
>       retryPolicy: { times: 2, delay: 300 },
>       set(key, value) { this[key] = value; }
>     };
>   }
>   return {
>     getInstance() {
>       if (!instance) instance = create();
>       return instance;
>     }
>   };
> })();
> 
> const a = RequestCenter.getInstance();
> const b = RequestCenter.getInstance();
> console.log(a === b); // true
> 
> 
> ```
> 阅读代码后，请选择错误的选项。

**A.** 闭包中的 `instance` 对外不可见，外部无法绕过 `getInstance()` 直接创建第二份配置。

**B.** `a === b` 为 `true`，说明多处调用拿到的是同一引用，策略修改可一处生效。

**C.** 单例意味着配置对象不可变，因此 `set` 方法不应存在，否则就破坏了单例语义。

**D.** 若未来需要多环境隔离（如测试/生产各一份），应评估是否仍适用单例，而非盲目套用。

**正确答案：** C　|　**我的答案：** C　|　✅

**答案详解：** 单例模式与配置对象的知识点：单例只约束实例数量唯一，并不要求对象不可变；唯一实例可保证多处修改一处生效。- ✅ A：闭包中的 instance 对外不可见，外部无法绕过 getInstance() 自行创建第二份配置，说法合理。- ✅ B：a === b 为 true 说明多次调用拿到同一引用，策略修改可一处生效，说法合理。- ❌ C：单例只保证“全局唯一实例”，并不意味配置对象不可变；配置对象带 set 方法（可变单例）很常见，可变性与唯一性并不冲突，“破坏单例语义”说法错误。- ✅ D：若未来需要测试/生产等多环境隔离，单例的全局唯一性会受限，应评估是否仍适用单例而非盲目套用，说法合理。

---

### 19. 全局 ThemeManager 需要保证全应用主题状态一致，应优先采用哪种实现？（5分）

> 背景
> 前端应用需要全局主题管理器。团队对比了「普通类多次 new」与「单例模式」两种写法。
> 实现方式一：普通类
> 
> ```js
> class ThemeManager {
>   constructor() {
>     this.theme = 'light';
>   }
>   setTheme(t) { this.theme = t; }
>   getTheme() { return this.theme; }
> }
> const a = new ThemeManager();
> const b = new ThemeManager();
> a.setTheme('dark');
> console.log(b.getTheme()); // 'light'
> console.log(a === b);       // false
> 
> 
> ```
> 实现方式二：单例
> 
> ```js
> class ThemeManager {
>   constructor() {
>     if (ThemeManager.instance) return ThemeManager.instance;
>     this.theme = 'light';
>     ThemeManager.instance = this;
>   }
>   static getInstance() {
>     if (!ThemeManager.instance) ThemeManager.instance = new ThemeManager();
>     return ThemeManager.instance;
>   }
> }
> const a = new ThemeManager();
> const b = new ThemeManager();
> a.setTheme('dark');
> console.log(b.getTheme()); // 'dark'
> console.log(a === b);       // true
> 
> 
> ```
> 全局主题状态需要一致，应优先采用？

**A.** 实现方式一，因为可以创建多个实例管理不同主题

**B.** 实现方式二，单例保证全局只有一个实例，适合管理全局共享状态

**C.** 两种实现完全等价，可随意互换

**D.** 实现方式二性能更差，应避免使用

**正确答案：** B　|　**我的答案：** B　|　✅

**答案详解：** 全局共享状态实现的知识点：多个主题实例会导致状态不一致，单例保证全局唯一实例，适合整体共享状态。- ❌ A：实现方式一可创建多个实例，多个 ThemeManager 会破坏全应用主题状态的一致性，不符合“全局主题一致”的要求。- ✅ B：实现方式二采用单例，全局只有一个实例，适合管理全局共享的主题状态，应优先采用。- ❌ C：两种实现语义不同，不能随意互换。- ❌ D：单例实现并没有显著的性能劣势，“性能更差应避免使用”说法错误。

---

### 20. 两种图表类型如何收口创建逻辑？（5分）

> 业务场景
> 运营后台数据看板根据 JSON 配置渲染图表，目前只有折线图（line）和柱状图（bar），PM 称两个季度内不会新增类型。业务代码曾散落 new LineChart(data)，构造函数一改就要动十几个文件。团队希望收口创建、调用方只 render()，且不为尚未到来的扩展堆叠过多抽象。
> 设计思想
> 简单工厂（Simple Factory） 由中心函数按参数创建不同产品，降低创建复杂度。产品种类少、变化慢时足够；类型持续增加时应再评估是否演进到工厂方法。
> 
> ```js
> function createChart(type, data) {
>   switch (type) {
>     case 'line': return new LineChart(data);
>     case 'bar':  return new BarChart(data);
>     default: throw new Error('unsupported chart');
>   }
> }
> 
> const chart = createChart('line', series);
> chart.render();
> 
> 
> ```
> 阅读代码后，请选择错误的选项。

**A.** 调用方只依赖 `createChart` 返回对象的 `render()`，与 `LineChart`/`BarChart` 解耦。

**B.** 当前仅两种图表且短期不新增时，为每种图单独建工厂类属于必要设计，否则无法扩展。

**C.** `default` 分支抛错，有助于在配置错误时快速失败，而不是静默返回 `undefined`。

**D.** 若 PM 宣布每月新增一种图表，应重新评估是否从简单工厂演进到工厂方法。

**正确答案：** B　|　**我的答案：** B　|　✅

**答案详解：** 简单工厂与工厂方法演进的知识点：规模小时单个创建函数足够、default 分支快速失败、规模增长后再演进复杂模式。- ✅ A：调用方只依赖 createChart 返回对象的 render() 方法，与 LineChart/BarChart 具体类解耦，说法合理。- ❌ B：当前仅两种图表且短期不新增时，用一个简单工厂即可满足需求；为每种图表单独建工厂类是过度设计（YAGNI），并非“不这样就无法扩展”，说法错误。- ✅ C：default 分支抛错可以让配置错误快速暴露（快速失败），而不是静默返回 undefined，说法合理。- ✅ D：若每月新增图表，简单工厂会被频繁改动，此时应重新评估是否演进到工厂方法等模式，说法合理。

---

