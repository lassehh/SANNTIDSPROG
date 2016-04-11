with Ada.Text_IO, Ada.Integer_Text_IO, Ada.Numerics.Float_Random;
use  Ada.Text_IO, Ada.Integer_Text_IO, Ada.Numerics.Float_Random;

procedure exercise8 is

    Count_Failed    : exception;    -- Exception to be raised when counting fails
    Gen             : Generator;    -- Random number generator

    protected type Transaction_Manager (N : Positive) is
        entry Finished;
        procedure Signal_Abort;
        entry Wait_Until_Aborted;
    private
        Finished_Gate_Open  : Boolean := False;
        Aborted             : Boolean := False;
    end Transaction_Manager;
    protected body Transaction_Manager is
        entry Finished when Finished_Gate_Open or Finished'Count = N is
        begin
        	if Finished'Count=N-1 then
        			Finished_Gate_Open := True;	
        	elsif Finished'Count = 0 then
        			Finished_Gate_Open := False;
        			Aborted :=False;
    		end if;
        end Finished;
        entry Wait_Until_Aborted when Aborted is
        begin
        	Aborted := True;
        end Wait_Until_Aborted;	



        procedure Signal_Abort is
        begin
            Aborted := True;
        end Signal_Abort;

    end Transaction_Manager;



    
    function Unreliable_Slow_Add (x : Integer) return Integer is
    Error_Rate : Constant := 0.15;  -- (between 0 and 1)
    random_number : float := Random(Gen);
    d : float := 4.0*Random(Gen);
    y : Integer;
    begin
    	if random_number < Error_Rate then
    		raise Count_Failed;
    	else
        	delay Duration(d);
        	y := x + 10;
        	return y;
        end if;   
    end Unreliable_Slow_Add;



    function Reliable_Add (x : Integer) return Integer is
     y : Integer;
    begin 
		y := x + 5;
		return y;
    end Reliable_Add;




    task type Transaction_Worker (Initial : Integer; Manager : access Transaction_Manager);
    task body Transaction_Worker is
        Num         : Integer   := Initial;
        Round_Num   : Integer   := 0;
        Prev 		: Integer 	:= Num;

    begin
    	
        Put_Line ("Worker" & Integer'Image(Initial) & " started");
		
        loop
            Put_Line ("Worker" & Integer'Image(Initial) & " started round" & Integer'Image(Round_Num));
            Round_Num := Round_Num + 1;	
			select
			    Manager.Wait_Until_Aborted;
			    Num:=Prev;
			    Num:=Reliable_Add(Num);
			    Put_Line ("  Worker" & Integer'Image(Initial) & " experienced a forward recovery, comitting " & Integer'Image(Num));
				Manager.Finished;
			then abort
				begin
					Num:=Unreliable_Slow_Add(Num);	
				exception
					when Count_Failed =>
						Manager.Signal_Abort;
				end;
			    Manager.Finished;
	            Put_Line ("  Worker" & Integer'Image(Initial) & " comitting" & Integer'Image(Num));
	            delay 0.5;
			end select;
			Prev := Num;
        end loop;
    end Transaction_Worker;

    Manager : aliased Transaction_Manager (3);

    Worker_1 : Transaction_Worker (0, Manager'Access);
    Worker_2 : Transaction_Worker (1, Manager'Access);
    Worker_3 : Transaction_Worker (2, Manager'Access);

begin
    Reset(Gen); -- Seed the random number generator
end exercise8;