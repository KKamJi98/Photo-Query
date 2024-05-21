# Amazon Photo Query - AI기반 사진 앨범 서비스

port : 8080

## Demo

- 자연어로 사진 검색  
[![자연어로 사진 검색](http://img.youtube.com/vi/jOgX3f43c1Q/0.jpg)](https://youtu.be/jOgX3f43c1Q)

- 얼굴 검색  
[![얼굴 검색](http://img.youtube.com/vi/jOgX3f43c1Q/0.jpg)](https://youtu.be/jOgX3f43c1Q)

- 태그 자동 생성  
[![태그 자동 생성](http://img.youtube.com/vi/GlHXMVzgk-s/0.jpg)](https://youtu.be/GlHXMVzgk-s)




## Complate :thumbsup:
- [x] Local Image S3 Upload
- [x] HTTP 요청을 통해 들어온 파일을 S3에 업로드
- [x] 사진 업로드 API 구현 [RDS에 데이터 적용, return msg]
- [x] 멀티파일 업로드 (zip파일 포함)
- [x] S3 Transfer Acceleration 적용
- [x] Go Routine 적용 
- [x] post 조회 API [RDS 생성 시 테이블에 맞게 수정]
- [x] post 생성 API 정리
- [x] post 삭제 API [RDS 생성 시 테이블에 맞게 수정]
- [x] CI/CD 파이프라인 구성 (Jenkins, ArgoCD)
- [x] 조회 무한 스크롤 지원
- [x] 북마크 API
- [x] 사진 삭제시 s3에서 original, thumbnail 사진까지 함께 삭제
- [x] dynamoDB에서 이미지 태그 데이터 불러오기
- [x] 모든 사진 삭제 API 생성
- [x] dynamoDB에서 불러온 모든 데이터에서 태그의 순위 구하기
- [x] 이미지 앨범 랭킹 기능 구현
## In Progress :fire:

## To Do :turtle:

## 아키텍처
### 클라우드 인프라
![image](https://github.com/War-Oxi/ACE-Team-KKamJi/assets/72260110/78b741a7-3437-4f64-b430-b0248175b9c0)

### 이미지 업로드 로직
![image](https://github.com/War-Oxi/ACE-Team-KKamJi/assets/72260110/5a51b26a-a296-4ac7-a436-a0112473c2d3)

## API 명세서
https://docs.google.com/spreadsheets/d/1b4K21uRSqM8BMv-PaZLhBoVLw22vt5KjUfqVjiu8_5k/edit#gid=0

## ERD 
https://www.erdcloud.com/d/RAvgh29cy7ZfpdeFt
