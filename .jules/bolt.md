## 2025-04-04 - BubbleTea TUI UI Lag
**Learning:** Found an $O(n^2)$ sorting bottleneck using custom nested loops for file sorting. In BubbleTea, long-running blocking tasks in update routines can cause the entire TUI to lock up.
**Action:** Replaced with O(n log n) standard library sorting. For TUI frameworks, always watch for O(n^2) logic handling dynamically sized inputs like directory listings to prevent UI freezing.
