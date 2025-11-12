import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatchOverPageComponent } from './match-over-page.component';


describe('MatchOverPageComponent', () => {
  let component: MatchOverPageComponent;
  let fixture: ComponentFixture<MatchOverPageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MatchOverPageComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(MatchOverPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
