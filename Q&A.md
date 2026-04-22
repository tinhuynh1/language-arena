# Q&A — Phỏng vấn Backend

---

## Bạn thích Golang ở điểm gì? So sánh với PHP.

Thú thật là em chưa có quá nhiều trải nghiệm sâu với PHP, nhưng qua tìm hiểu, em thấy mỗi ngôn ngữ đều có thế mạnh riêng. Em chọn Go vì 3 lý do:

- **Sự an toàn** — Go rất chặt chẽ về kiểu dữ liệu. Điều này giúp em kiểm soát code tốt hơn, ít lỗi vặt khi hệ thống mở rộng và tự tin hơn mỗi khi cần chỉnh sửa hay nâng cấp.
- **Tối ưu tài nguyên** — Go cực mạnh trong việc xử lý nhiều tác vụ cùng lúc mà không tốn quá nhiều phần cứng. Với các hệ thống cần tốc độ phản hồi nhanh, Go giúp tiết kiệm chi phí vận hành hơn nhiều.
- **Hệ sinh thái hiện đại** — Hiện nay các công cụ hạ tầng lớn (như Docker hay Kubernetes) đều dùng Go. Việc giỏi ngôn ngữ này giúp em dễ dàng làm chủ công nghệ và xây dựng các hệ thống Microservices linh hoạt.

Tóm lại, nếu PHP mạnh về việc đưa sản phẩm ra thị trường nhanh, thì Go phù hợp để xây dựng những hệ thống có hiệu năng cao và ổn định lâu dài.

---

## Bạn thích giải quyết những vấn đề gì nhất trong backend?

Em thích xây dựng các hệ thống có tính sẵn sàng cao và khả năng mở rộng linh hoạt. Thay vì chỉ tập trung vào code, em đặc biệt chú trọng đến việc tối ưu hóa luồng dữ liệu để hệ thống luôn vận hành mượt mà, ngay cả trong những khung giờ cao điểm.

Cụ thể, em thường tiếp cận vấn đề theo 3 lớp:

- **Tối ưu từ gốc** — Kiểm soát chặt chẽ các truy vấn Database, xử lý dứt điểm các lỗi kinh điển như N+1 hay thiếu Index để đảm bảo tài nguyên không bị nghẽn.
- **Chiến lược phân tải** — Áp dụng Sharding, Partitioning hoặc Caching đa tầng (Redis, CDN) để giảm áp lực cho hệ thống lõi và tăng tốc độ phản hồi.
- **Sự ổn định toàn diện** — Thiết kế các service giao tiếp với nhau hiệu quả, đảm bảo tính ổn định và khả năng phục hồi của hệ thống trước lượng truy vấn lớn.

Tại công ty cũ, em từng tối ưu thành công các API bị chậm bằng cách analyze lại câu lệnh SQL và tái cấu trúc logic để tránh truy vấn thừa, giúp cải thiện đáng kể trải nghiệm người dùng.

---

## Bạn thấy vấn đề gì khó nhất trong backend?

Theo quan điểm của em, vấn đề khó nhất trong Backend là việc đưa ra các quyết định đánh đổi (Trade-offs) dựa trên bối cảnh thực tế.

- **Xác định mức độ ưu tiên (định lý CAP)** — Tùy vào nghiệp vụ mà mình phải hy sinh một yếu tố để đạt được yếu tố khác. Ví dụ, với hệ thống tài chính, em ưu tiên sự Nhất quán (Consistency); nhưng với các hệ thống cần tương tác nhanh, em sẽ ưu tiên sự Sẵn sàng (Availability).
- **Quản lý rủi ro khi mở rộng** — Mỗi lựa chọn thiết kế đều có hệ quả lâu dài. Một quyết định sai lầm ở giai đoạn thiết kế có thể gây tắc nghẽn hoặc mất dữ liệu khi lượng người dùng tăng đột biến.
- **Tìm giải pháp trung hòa** — Thử thách thực sự là tìm ra điểm cân bằng phù hợp nhất, ví dụ như áp dụng Nhất quán sau cùng (Eventual Consistency) để vừa đảm bảo trải nghiệm người dùng mượt mà, vừa giữ được độ tin cậy cho dữ liệu.

Tóm lại, em nghĩ việc hiểu rõ cái giá phải trả cho mỗi giải pháp kỹ thuật là điều khó nhất và cũng quan trọng nhất đối với một lập trình viên Backend.

---

## Hãy kể về một tình huống trong dự án mà bạn tự hào khi giải quyết được.

Ở công ty cũ, em từng xử lý một sự cố xảy ra vào đúng khung giờ cao điểm. Lượng truy cập tăng đột biến khiến hệ thống phản hồi chậm, người dùng liên tục nhấn lại nút đặt hàng, dẫn đến tình trạng trùng đơn trên diện rộng — ảnh hưởng trực tiếp đến vận hành và uy tín công ty.

Áp lực lúc đó rất lớn vì thiệt hại đang tăng lên từng phút. Nhưng em xác định rằng nếu vội vàng xử lý mà không tìm đúng nguyên nhân, rủi ro gây ra sự cố nặng hơn là rất cao.

Thay vì hành động ngay, em dành thời gian kiểm tra log và dữ liệu để xác định chính xác vấn đề: hệ thống bị race condition khi nhiều request cùng ghi vào một bản ghi đơn hàng đồng thời. Sau khi có đủ bằng chứng, em đề xuất với team áp dụng distributed lock bằng Redis để đảm bảo mỗi đơn hàng chỉ được xử lý bởi một request tại một thời điểm, mà không làm tăng độ trễ đáng kể cho người dùng.

Sau khi triển khai, tình trạng trùng đơn chấm dứt hoàn toàn và hệ thống ổn định trở lại. Điều em tự hào nhất không chỉ là kết quả kỹ thuật, mà là đã giữ được sự điềm tĩnh để đưa ra quyết định đúng đắn dưới áp lực.